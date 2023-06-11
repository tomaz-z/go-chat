package connection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nhooyr.io/websocket"
)

func TestNew(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, New())
}

type connection struct {
	closed bool
	err    bool
}

func (c *connection) Close(code websocket.StatusCode, reason string) error {
	c.closed = true

	if c.err {
		return assert.AnError
	}

	return nil
}

func TestAdd(t *testing.T) {
	t.Parallel()

	s := &service{
		connections: []Connection{},
	}

	s.Add(&connection{})

	assert.Equal(t, 1, len(s.connections))

	s.Add(&connection{})

	assert.Equal(t, 2, len(s.connections))
}

func TestClose(t *testing.T) {
	t.Parallel()

	c1 := &connection{
		err: false,
	}

	c2 := &connection{
		err: true,
	}

	s := &service{
		connections: []Connection{c1, c2},
	}

	s.Close()

	assert.True(t, c1.closed)
	assert.True(t, c2.closed)
}
