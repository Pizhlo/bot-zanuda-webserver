package v0

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type test struct {
		name  string
		space spaceService
		user  userService
		auth  authService
		err   error
		want  *handler
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceSrvMock, userSrvMock, authSrvMock := createMockServices(ctrl)

	tests := []test{
		{
			name:  "success",
			space: spaceSrvMock,
			user:  userSrvMock,
			auth:  authSrvMock,
			err:   nil,
			want:  &handler{space: spaceSrvMock, user: userSrvMock, auth: authSrvMock},
		},
		{
			name:  "space is nil",
			space: nil,
			user:  userSrvMock,
			auth:  authSrvMock,
			err:   errors.New("space is nil"),
			want:  nil,
		},
		{
			name:  "user is nil",
			space: spaceSrvMock,
			user:  nil,
			auth:  authSrvMock,
			err:   errors.New("user is nil"),
			want:  nil,
		},
		{
			name:  "auth is nil",
			space: spaceSrvMock,
			user:  userSrvMock,
			auth:  nil,
			err:   errors.New("auth is nil"),
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := New(tt.space, tt.user, tt.auth)
			if tt.err != nil {
				require.Error(t, err)
				require.Equal(t, tt.err, err)
				require.Nil(t, handler)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, handler)
			}
		})
	}
}
