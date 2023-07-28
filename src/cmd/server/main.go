package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	chatAPI "gochat/cmd/server/handlers/chat"
	joinAPI "gochat/cmd/server/handlers/join"
	"gochat/internal/chat"
	"gochat/internal/storage/inmemory/user"
	"gochat/internal/websocket/connection"
)

const EnvGoChatPort = "GO_CHAT_PORT"

func main() {
	port := os.Getenv(EnvGoChatPort)
	if len(port) < 1 {
		port = "4001"
	}

	log.Println("Starting server...")

	userStorage := user.New()

	connService := connection.New()
	chatService := chat.New()

	joinHandler := joinAPI.New(userStorage)
	chatHandler := chatAPI.New(userStorage, connService, chatService)

	http.HandleFunc("/join", joinHandler.Join)
	http.HandleFunc("/subscribe", chatHandler.Subscribe)
	http.HandleFunc("/publish", chatHandler.Publish)

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM)

	srv := http.Server{
		Addr: fmt.Sprintf(":%s", port),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error on serving API: %v\n", err)
		}
	}()

	log.Printf("Server ready on %s!\n", srv.Addr)

	<-term

	log.Println("Closing websocket connections...")

	connService.Close()

	log.Println("Websocket connections closed")

	log.Println("Stopping server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("error shutting down server: %v\n", err)
	}

	log.Println("Server stopped.")
}
