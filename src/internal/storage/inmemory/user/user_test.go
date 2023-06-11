package user

import (
	"errors"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, New())
}

func TestUserStorage(t *testing.T) {
	t.Parallel()

	s := &storage{
		users: map[ulid.ULID]string{},
	}

	mockToken := ulid.MustParse("01H2NEEJ6ZVYSBEENV616A1RZ7")
	mockUser := "user"

	assert.Empty(t, s.Get(mockToken))

	u, err := s.FindTokenByUsername(mockUser)
	assert.Equal(t, errors.New("not found"), err)
	assert.Empty(t, u)

	s.Set(mockToken, mockUser)

	assert.Equal(t, mockUser, s.Get(mockToken))

	u, err = s.FindTokenByUsername(mockUser)
	assert.NoError(t, err)
	assert.Equal(t, mockToken, u)

	s.Remove(mockToken)

	assert.Empty(t, s.Get(mockToken))
}
