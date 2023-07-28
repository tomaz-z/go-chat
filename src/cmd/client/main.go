package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/oklog/ulid/v2"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	BearerToken         = "Bearer"
	EnvGoChatServerHost = "GO_CHAT_SERVER_HOST"
)

type User struct {
	Name string
}

func (t User) JSON() ([]byte, error) {
	return json.Marshal(t)
}

type Token struct {
	Value ulid.ULID
}

type GreetMessage struct {
	Token ulid.ULID
}

type Message struct {
	Author string
	Value  string
}

func main() {
	host := os.Getenv(EnvGoChatServerHost)
	if len(host) < 1 {
		host = "localhost:4001"
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("User name: ")
	author, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err, "error getting message")
	}
	author = strings.TrimSpace(author)

	user := User{
		Name: author,
	}

	userJson, err := user.JSON()
	if err != nil {
		log.Fatal(err, "error marshalling user")
	}

	log.Println("Joining chat and obtaining token...")

	res, err := http.Post(fmt.Sprintf("http://%s/join", host), "application/json", bytes.NewReader(userJson))
	if err != nil {
		log.Fatal(err, "error on POST /join")
	}

	if res.StatusCode > 399 {
		log.Fatal(err, "non-200 response returned")
	}

	tokenBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err, "error reading token from body")
	}

	var token Token
	err = json.Unmarshal(tokenBody, &token)
	if err != nil {
		log.Fatal(err, "error unmarshalling token")
	}

	log.Println("Token received!")

	// Subscribe to messages.
	var cSubscribe *websocket.Conn
	go func() {
		cSubscribe, _, err = websocket.Dial(context.Background(), fmt.Sprintf("ws://%s/subscribe", host), &websocket.DialOptions{
			HTTPHeader: http.Header{BearerToken: []string{token.Value.String()}},
		})
		if err != nil {
			log.Fatal(err, "error dialing server")
		}
		defer cSubscribe.Close(websocket.StatusInternalError, "the sky is falling")

		log.Println("Subscribe: OK")

		for {
			var msg Message
			err = wsjson.Read(context.Background(), cSubscribe, &msg)
			if err != nil {
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure || errors.Is(err, context.Canceled) {
					// On server closure, termination should happen on clients.
					interrupt <- syscall.SIGTERM

					return
				}

				log.Fatal(err, "error receiving message")
			}

			log.Printf("%s: %s", msg.Author, msg.Value)
		}
	}()

	// Publish messages.
	var cPublish *websocket.Conn
	go func() {
		cPublish, _, err = websocket.Dial(context.Background(), fmt.Sprintf("ws://%s/publish", host), &websocket.DialOptions{
			HTTPHeader: http.Header{BearerToken: []string{token.Value.String()}},
		})
		if err != nil {
			log.Fatal(err, "error dialing server")
		}
		defer cPublish.Close(websocket.StatusInternalError, "the sky is falling")

		log.Println("Publish: OK")

		for {
			reader := bufio.NewReader(os.Stdin)
			message, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err, "error getting message")
			}
			message = strings.TrimSpace(message)

			ctxMessage, cancelMessage := context.WithTimeout(context.Background(), time.Second*10)
			err = wsjson.Write(ctxMessage, cPublish, Message{
				Author: author,
				Value:  message,
			})
			cancelMessage()
			if err != nil {
				log.Fatal(err, "error sending message")
			}
		}
	}()

	<-interrupt

	log.Println("Stopping client...")

	if err := cSubscribe.Close(websocket.StatusNormalClosure, "stopping subscribe client"); err != nil {
		log.Println(err, "error closing subscribe websocket client")
	}

	if err := cPublish.Close(websocket.StatusNormalClosure, "stopping publish server"); err != nil {
		log.Println(err, "error closing publish websocket client")
	}

	log.Println("Client stopped.")
}
