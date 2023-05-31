package chat

import (
	"context"
	"log"
	"net/http"

	"github.com/oklog/ulid/v2"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"gochat/cmd/server/models"
	"gochat/internal/storage/inmemory/user"
	"gochat/internal/websocket/connection"
)

type ChatHandler interface {
	Publish(w http.ResponseWriter, r *http.Request)
	Subscribe(w http.ResponseWriter, r *http.Request)
}

func New(userStorage user.UserStorage, connService connection.ConnectionService) ChatHandler {
	return handler{
		userStorage: userStorage,
		connService: connService,
	}
}

// func announce(connections map[ulid.ULID]*websocket.Conn, msg Message) {
// 	for _, conn := range connections {
// 		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

// 		err := wsjson.Write(ctx, conn, msg)
// 		cancel()
// 		if err != nil {
// 			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
// 				return
// 			}

// 			log.Fatal(err, "error sending message")
// 		}
// 	}
// }

type handler struct {
	userStorage user.UserStorage
	connService connection.ConnectionService
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

	h.connService.Set(token, c)

	// announce - subscribe should listen to channel? would that work?
	// announce(connections, Message{
	// 	Author: GoChatName,
	// 	Value:  fmt.Sprintf("%s has joined the chat!", user),
	// })

	for {
		var msg models.Message
		err := wsjson.Read(context.Background(), c, &msg)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				h.connService.Remove(token)
				h.userStorage.Remove(token)

				// announce(connections, Message{
				// 	Author: GoChatName,
				// 	Value:  fmt.Sprintf("%s has left the chat!", user),
				// })

				return
			}

			log.Fatal(err, "error reading message")
		}

		// announce(connections, msg)
	}
}

func (h handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	// todo just listen to a channel?
}
