package v0

import (
	"webserver/internal/service/space"
	"webserver/internal/service/user"
)

type Handler struct {
	space *space.Space
	user  *user.User
}

func New(space *space.Space, user *user.User) *Handler {
	return &Handler{space: space, user: user}
}
