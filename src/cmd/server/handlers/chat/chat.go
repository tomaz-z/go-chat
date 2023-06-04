package chat

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"gochat/cmd/server/models"
	"gochat/internal/chat"
	"gochat/internal/storage/inmemory/user"
	"gochat/internal/websocket/connection"
)

type ChatHandler interface {
	Publish(w http.ResponseWriter, r *http.Request)
	Subscribe(w http.ResponseWriter, r *http.Request)
}

func New(
	userStorage user.UserStorage,
	connService connection.ConnectionService,
	chatService chat.ChatService,
) ChatHandler {
	return handler{
		userStorage: userStorage,
		connService: connService,
		chatService: chatService,
	}
}

type handler struct {
	userStorage user.UserStorage
	connService connection.ConnectionService
	chatService chat.ChatService
}

func (h handler) Publish(w http.ResponseWriter, r *http.Request) {
	tokenHeader := r.Header.Get(models.BearerToken)
	token, err := ulid.Parse(tokenHeader)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	user := h.userStorage.Get(token)
	if len(user) < 1 {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Fatal(err, "error getting connection")

		return
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	//h.connService.Set(token, c) //TODO remove connService concept?

	// Announce new user.
	h.chatService.PostMessage(chat.Message{
		Author:  chat.ChatAPIName,
		Message: fmt.Sprintf("%s has joined the chat!", user),
	})

	for {
		var msg models.Message
		err := wsjson.Read(context.Background(), c, &msg)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				h.connService.Remove(token)
				h.userStorage.Remove(token)

				// Announce user left.
				h.chatService.PostMessage(chat.Message{
					Author:  chat.ChatAPIName,
					Message: fmt.Sprintf("%s has left the chat!", user),
				})

				return
			}

			log.Fatal(err, "error reading message")
		}

		// Announce user message.
		h.chatService.PostMessage(chat.Message{
			Author:  msg.Author,
			Message: msg.Value,
		})
	}
}

func (h handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	tokenHeader := r.Header.Get(models.BearerToken)
	token, err := ulid.Parse(tokenHeader)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	user := h.userStorage.Get(token)
	if len(user) < 1 {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Fatal(err, "error getting connection")

		return
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	// TODO: properly handle open connections

	for msg := range h.chatService.Subscribe() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

		err := wsjson.Write(ctx, c, models.Message{
			Author: msg.Author,
			Value:  msg.Message,
		})
		cancel()
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				return
			}

			log.Fatal(err, "error sending message")
		}
	}
}
