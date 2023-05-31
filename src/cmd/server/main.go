package main

import (
	"context"
	"encoding/json"
	"errors"
	"gochat/cmd/server/handlers/chat"
	"gochat/cmd/server/handlers/join"
	"gochat/internal/storage/inmemory/user"
	"gochat/internal/websocket/connection"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/oklog/ulid/v2"
)

const (
	GoChatName = "GoChat"
)

type Token struct {
	Value ulid.ULID
}

func (t Token) JSON() ([]byte, error) {
	return json.Marshal(t)
}

type GreetMessage struct {
	Token ulid.ULID
}

func main() {
	log.Println("Starting server...")

	userStorage := user.New()

	connService := connection.New()

	joinHandler := join.New(userStorage)
	chatHandler := chat.New(userStorage, connService)

	http.HandleFunc("/join", joinHandler.Join)
	http.HandleFunc("/publish", chatHandler.Publish)
	http.HandleFunc("/subscribe", chatHandler.Subscribe)

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM)

	srv := http.Server{
		Addr: ":4001",
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err, "error on serving API")
		}
	}()

	log.Println("Server ready!")

	<-term

	log.Println("Closing open websocket connections...")

	connService.Close()

	log.Println("Stopping server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatal(err, "error shutting down server")
	}

	log.Println("Server stopped.")
}
