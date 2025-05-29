package v0

import (
	"testing"
	"webserver/internal/service/auth"
	"webserver/internal/service/space"
	"webserver/internal/service/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	space := space.New(nil, nil, nil)
	user := user.New(nil, nil)
	authCfg, err := auth.NewConfig([]byte{123})
	require.NoError(t, err)

	auth, err := auth.New(authCfg)
	require.NoError(t, err)

	expectedH := &handler{
		space: space,
		user:  user,
		auth:  auth,
	}

	handler := New(space, user, auth)

	assert.Equal(t, space, handler.space)
	assert.Equal(t, user, handler.user)
	assert.Equal(t, expectedH, handler)
}
