package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type Message struct {
	Author string
	Value  string
}

func main() {
	log.Println("Starting server...")

	http.Handle("/messages", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Fatal(err, "error getting connection")
			return
		}
		defer c.Close(websocket.StatusNormalClosure, "")

		for {
			ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
			defer cancel()

			var v Message
			err = wsjson.Read(ctx, c, &v)
			if err != nil {
				log.Fatal(err, "error getting connection")
				return
			}

			log.Printf("%s: %s", v.Author, v.Value)
		}
	}))

	log.Println("Server ready!")

	if err := http.ListenAndServe(":4001", nil); err != nil {
		log.Fatal(err, "error on serving API")
	}
}
