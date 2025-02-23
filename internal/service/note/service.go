package note

import (
	"context"
	"webserver/internal/model"
)

// Note обрабатывает запросы о заметках - создание, удаление, редактирование, получение
type Note struct {
	repo repo
}

//go:generate mockgen -source ./service.go -destination=../../../mocks/note_srv.go -package=mocks
type repo interface {
	Create(ctx context.Context, note model.CreateNoteRequest) error
}

func New(repo repo) *Note {
	return &Note{repo: repo}
}

// Create создает заметку пользователя в БД
func (s *Note) Create(ctx context.Context, note model.CreateNoteRequest) error {
	return s.repo.Create(ctx, note)
}
