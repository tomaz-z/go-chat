package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type Message struct {
	Author string
	Value  string
}

func main() {
	c, _, err := websocket.Dial(context.Background(), "ws://localhost:4001/messages", nil)
	if err != nil {
		log.Fatal(err, "error dialing server")
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("User name: ")
	author, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err, "error getting message")
	}
	author = strings.TrimSpace(author)

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Message: ")
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
