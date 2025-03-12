package space

import (
	"context"
	"webserver/internal/model"
)

type Space struct {
	repo spaceRepo
}

//go:generate mockgen -source ./space.go -destination=../../../mocks/space_srv.go -package=mocks
type spaceRepo interface {
	GetSpaceByID(ctx context.Context, id int) (model.Space, error)
	// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает в полном виде.
	GetAllbySpaceIDFull(ctx context.Context, spaceID int64) ([]model.Note, error)
	// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает кратко (только userID)
	GetAllBySpaceID(ctx context.Context, spaceID int64) ([]model.GetNote, error)
}

func New(repo spaceRepo) *Space {
	return &Space{repo: repo}
}

func (s *Space) GetSpaceByID(ctx context.Context, id int) (model.Space, error) {
	return s.repo.GetSpaceByID(ctx, id)
}

// GetAllbySpaceIDFull возвращает все заметки пространства.
// Информацию о пользователе возвращает в полном виде.
func (s *Space) GetAllbySpaceIDFull(ctx context.Context, spaceID int64) ([]model.Note, error) {
	return s.repo.GetAllbySpaceIDFull(ctx, spaceID)
}

// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает кратко (только userID)
func (s *Space) GetAllBySpaceID(ctx context.Context, spaceID int64) ([]model.GetNote, error) {
	return s.repo.GetAllBySpaceID(ctx, spaceID)
}
