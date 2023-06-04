package connection

import (
	"log"

	"nhooyr.io/websocket"
)

type ConnectionService interface {
	Add(conn *websocket.Conn)
	Close()
}

func New() ConnectionService {
	return &service{
		connections: []*websocket.Conn{},
	}
}

type service struct {
	connections []*websocket.Conn
}

func (s *service) Add(conn *websocket.Conn) {
	s.connections = append(s.connections, conn)
}

func (s *service) Close() {
	for _, c := range s.connections {
		if err := c.Close(websocket.StatusNormalClosure, "stopping server"); err != nil {
			log.Println(err, "error closing websocket client")
		}
	}
}
