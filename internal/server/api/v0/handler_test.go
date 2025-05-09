package v0

import (
	"testing"
	"webserver/internal/service/space"
	"webserver/internal/service/user"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	space := space.New(nil, nil, nil)
	user := user.New(nil, nil)

	expectedH := &handler{
		space: space,
		user:  user,
	}

	handler := New(space, user)

	assert.Equal(t, space, handler.space)
	assert.Equal(t, user, handler.user)
	assert.Equal(t, expectedH, handler)
}
