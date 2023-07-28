package connection

import (
	"log"

	"nhooyr.io/websocket"
)

type Connection interface {
	Close(code websocket.StatusCode, reason string) error
}

type ConnectionService interface {
	Add(conn Connection)
	Close()
}

func New() ConnectionService {
	return &service{
		connections: []Connection{},
	}
}

type service struct {
	connections []Connection
}

func (s *service) Add(conn Connection) {
	s.connections = append(s.connections, conn)
}

func (s *service) Close() {
	for _, c := range s.connections {
		if err := c.Close(websocket.StatusNormalClosure, "stopping server"); err != nil {
			log.Printf("error closing websocket client: %v\n", err)
		}
	}
}
