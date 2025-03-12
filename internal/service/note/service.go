package note

import (
	"context"
	"webserver/internal/model"
)

// Note обрабатывает запросы о заметках - создание, удаление, редактирование, получение
type Note struct {
	repo noteRepo
}

//go:generate mockgen -source ./service.go -destination=../../../mocks/note_srv.go -package=mocks
type noteRepo interface {
	// Create создает новую заметку в личном пространстве пользователя
	Create(ctx context.Context, note model.CreateNoteRequest) error
}

func New(repo noteRepo) *Note {
	return &Note{repo: repo}
}

// Create создает новую заметку в личном пространстве пользователя
func (s *Note) Create(ctx context.Context, note model.CreateNoteRequest) error {
	return s.repo.Create(ctx, note)
}
