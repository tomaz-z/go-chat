package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, New())
}

func TestPostMessage(t *testing.T) {
	t.Parallel()

	s := &service{
		subscriptions: []chan Message{make(chan Message, 1)},
	}

	s.PostMessage(Message{})

	assert.NotEmpty(t, s.subscriptions[0])
}

func TestSubscribe(t *testing.T) {
	t.Parallel()

	s := &service{
		subscriptions: []chan Message{},
	}

	assert.NotNil(t, s.Subscribe())

	assert.Equal(t, 1, len(s.subscriptions))
}
