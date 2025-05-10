package v0

import (
	"webserver/internal/service/auth"
	"webserver/internal/service/space"
	"webserver/internal/service/user"
)

type handler struct {
	space *space.Space
	user  *user.User
	auth  *auth.AuthService
}

func New(space *space.Space, user *user.User, auth *auth.AuthService) *handler {
	return &handler{space: space, user: user, auth: auth}
}
