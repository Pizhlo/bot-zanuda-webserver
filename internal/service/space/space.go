package space

import (
	"context"
	"webserver/internal/model"
)

type Space struct {
	repo repo
}

type repo interface {
	GetSpaceByID(ctx context.Context, id int) (model.Space, error)
}

func New(repo repo) *Space {
	return &Space{repo: repo}
}

func (s *Space) GetSpaceByID(ctx context.Context, id int) (model.Space, error) {
	return s.repo.GetSpaceByID(ctx, id)
}
