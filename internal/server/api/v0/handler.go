package v0

import (
	"webserver/internal/service/space"
	"webserver/internal/service/user"
)

type handler struct {
	space *space.Space
	user  *user.User
}

func New(space *space.Space, user *user.User) *handler {
	return &handler{space: space, user: user}
}
