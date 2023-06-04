package user

import (
	"errors"
	"sync"

	"github.com/oklog/ulid/v2"
)

type UserStorage interface {
	FindTokenByUsername(username string) (ulid.ULID, error)
	Set(token ulid.ULID, username string)
	Get(token ulid.ULID) string
	Remove(token ulid.ULID)
}

func New() UserStorage {
	return &storage{
		users: map[ulid.ULID]string{},
	}
}

type storage struct {
	sync.Mutex
	users map[ulid.ULID]string
}

func (s *storage) FindTokenByUsername(username string) (ulid.ULID, error) {
	s.Lock()
	defer s.Unlock()

	for k, u := range s.users {
		if u == username {
			return k, nil
		}
	}

	return ulid.ULID{}, errors.New("not found")
}

func (s *storage) Set(token ulid.ULID, username string) {
	s.Lock()
	defer s.Unlock()

	s.users[token] = username
}

func (s *storage) Get(token ulid.ULID) string {
	s.Lock()
	defer s.Unlock()

	return s.users[token]
}

func (s *storage) Remove(token ulid.ULID) {
	s.Lock()
	defer s.Unlock()

	delete(s.users, token)
}
