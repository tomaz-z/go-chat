package connection

import (
	"log"
	"sync"

	"github.com/oklog/ulid/v2"
	"nhooyr.io/websocket"
)

type ConnectionService interface {
	Set(token ulid.ULID, conn *websocket.Conn)
	Get(token ulid.ULID) *websocket.Conn
	Remove(token ulid.ULID)
	Close()
}

func New() ConnectionService {
	return &service{
		connections: map[ulid.ULID]*websocket.Conn{},
	}
}

type service struct {
	*sync.Mutex
	connections map[ulid.ULID]*websocket.Conn
}

func (s *service) Set(token ulid.ULID, conn *websocket.Conn) {
	s.Lock()
	defer s.Unlock()

	s.connections[token] = conn
}

func (s *service) Get(token ulid.ULID) *websocket.Conn {
	s.Lock()
	defer s.Unlock()

	return s.connections[token]
}

func (s *service) Remove(token ulid.ULID) {
	s.Lock()
	defer s.Unlock()

	delete(s.connections, token)
}

func (s *service) Close() {
	s.Lock()
	defer s.Unlock()

	for _, c := range s.connections {
		if err := c.Close(websocket.StatusNormalClosure, "stopping server"); err != nil {
			log.Println(err, "error closing websocket client")
		}
	}
}
