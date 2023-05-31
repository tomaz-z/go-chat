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

	res, err := http.Post("http://localhost:4001/join", "application/json", bytes.NewReader(userJson))
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

	c, _, err := websocket.Dial(context.Background(), "ws://localhost:4001/chat", &websocket.DialOptions{
		// HTTPHeader: ,
	})
	if err != nil {
		log.Fatal(err, "error dialing server")
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	ctx := context.Background()

	ctxGreet, cancelGreet := context.WithTimeout(ctx, time.Second*10)
	err = wsjson.Write(ctxGreet, c, GreetMessage{
		Token: token.Value,
	})
	cancelGreet()
	if err != nil {
		log.Fatal(err, "error sending message")
	}

	go func() {
		for {
			var msg Message
			err = wsjson.Read(ctx, c, &msg)
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

	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
			message, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err, "error getting message")
			}
			message = strings.TrimSpace(message)

			ctxMessage, cancelMessage := context.WithTimeout(ctx, time.Second*10)
			err = wsjson.Write(ctxMessage, c, Message{
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

	log.Println("Stopping websocket client...")

	if err = c.Close(websocket.StatusNormalClosure, "stopping client"); err != nil {
		log.Println(err, "error closing websocket client")

		return
	}

	log.Println("Websocket client closed!")
}
