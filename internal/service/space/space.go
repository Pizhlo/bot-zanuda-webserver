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
	// CreateNote создает новую заметку в пространстве пользователя
	CreateNote(ctx context.Context, note model.CreateNoteRequest) error
	GetSpaceByID(ctx context.Context, id int) (model.Space, error)
	// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает в полном виде.
	GetAllNotesBySpaceIDFull(ctx context.Context, spaceID int64) ([]model.Note, error)
	// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает кратко (только userID)
	GetAllNotesBySpaceID(ctx context.Context, spaceID int64) ([]model.GetNote, error)
	UpdateNote(ctx context.Context, update model.UpdateNote) error
}

func New(repo spaceRepo) *Space {
	return &Space{repo: repo}
}

func (s *Space) GetSpaceByID(ctx context.Context, id int) (model.Space, error) {
	return s.repo.GetSpaceByID(ctx, id)
}

// Create создает новую заметку в личном пространстве пользователя
func (s *Space) CreateNote(ctx context.Context, note model.CreateNoteRequest) error {
	err := note.Validate()
	if err != nil {
		return err
	}

	return s.repo.CreateNote(ctx, note)
}

// GetAllbySpaceIDFull возвращает все заметки пространства.
// Информацию о пользователе возвращает в полном виде.
func (s *Space) GetAllbySpaceIDFull(ctx context.Context, spaceID int64) ([]model.Note, error) {
	return s.repo.GetAllNotesBySpaceIDFull(ctx, spaceID)
}

// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает кратко (только userID)
func (s *Space) GetAllBySpaceID(ctx context.Context, spaceID int64) ([]model.GetNote, error) {
	return s.repo.GetAllNotesBySpaceID(ctx, spaceID)
}

func (s *Space) UpdateNote(ctx context.Context, update model.UpdateNote) error {
	err := update.Validate()
	if err != nil {
		return err
	}

	return s.repo.UpdateNote(ctx, update)
}
