package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	GoChatName = "GoChat"
)

type User struct {
	Name string
}

type Token struct {
	Value ulid.ULID
}

func (t Token) JSON() ([]byte, error) {
	return json.Marshal(t)
}

type GreetMessage struct {
	Token ulid.ULID
}

type Message struct {
	Author string
	Value  string
}

func announce(connections map[ulid.ULID]*websocket.Conn, msg Message) {
	for _, conn := range connections {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

		err := wsjson.Write(ctx, conn, msg)
		cancel()
		if err != nil {
			log.Fatal(err, "error sending message")
		}
	}
}

func main() {
	log.Println("Starting server...")

	chatHistory := []Message{}
	users := map[ulid.ULID]string{}
	connections := map[ulid.ULID]*websocket.Conn{}

	http.Handle("/join", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err, "error reading body")
		}

		var user User
		if err := json.Unmarshal(body, &user); err != nil {
			log.Fatal(err, "error unmarshalling body")
		}

		token := ulid.Make()

		if _, ok := users[token]; ok {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		users[token] = user.Name

		tokenJson, err := Token{
			Value: token,
		}.JSON()
		if err != nil {
			log.Fatal(err, "error marshalling token")
		}

		w.Write(tokenJson)
	}))

	http.Handle("/chat", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Fatal(err, "error getting connection")
			return
		}
		defer c.Close(websocket.StatusNormalClosure, "")

		var msg GreetMessage
		err = wsjson.Read(context.Background(), c, &msg)
		if err != nil {
			log.Fatal(err, "error getting connection")
			return
		}

		user, ok := users[msg.Token]
		if !ok {
			log.Fatal("invalid token")
			return
		}

		connections[msg.Token] = c

		announce(connections, Message{
			Author: GoChatName,
			Value:  fmt.Sprintf("%s has joined the chat!", user),
		})

		for {
			var msg Message
			err = wsjson.Read(context.Background(), c, &msg)
			if err != nil {
				log.Fatal(err, "error getting connection")
				return
			}

			chatHistory = append(chatHistory, msg)

			announce(connections, msg)
		}
	}))

	log.Println("Server ready!")

	if err := http.ListenAndServe(":4001", nil); err != nil {
		log.Fatal(err, "error on serving API")
	}
}
