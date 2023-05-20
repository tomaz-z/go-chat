package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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

	c, _, err := websocket.Dial(context.Background(), "ws://localhost:4001/chat", nil)
	if err != nil {
		log.Fatal(err, "error dialing server")
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	err = wsjson.Write(ctx, c, GreetMessage{
		Token: token.Value,
	})
	cancel()
	if err != nil {
		log.Fatal(err, "error sending message")
	}

	go func() {
		for {
			var msg Message
			err = wsjson.Read(context.Background(), c, &msg)
			if err != nil {
				log.Fatal(err, "error sending message")
			}

			log.Printf("%s: %s", msg.Author, msg.Value)
		}
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err, "error getting message")
		}
		message = strings.TrimSpace(message)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		err = wsjson.Write(ctx, c, Message{
			Author: author,
			Value:  message,
		})
		cancel()
		if err != nil {
			log.Fatal(err, "error sending message")
		}
	}
}
