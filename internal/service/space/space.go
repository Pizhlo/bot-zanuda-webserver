package space

import (
	"context"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"

	"github.com/google/uuid"
)

type Space struct {
	repo  spaceRepo
	cache cache
	saver dbWorker // создание / обновление записей
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
	// CheckParticipant проверяет, является ли пользователь участником пространства
	CheckParticipant(ctx context.Context, userID int64, spaceID uuid.UUID) error
	// CheckIfNoteExistsInSpace проверяет, что в пространстве существует такая заметка
	CheckIfNoteExistsInSpace(ctx context.Context, noteID, spaceID uuid.UUID) error
}

type cache interface {
	// CheckParticipant проверяет, является ли пользователь участником пространства
	CheckParticipant(ctx context.Context, userID int64, spaceID uuid.UUID) error
}

// dbWorker работает на создание / обновление записей
type dbWorker interface {
	AddQuery(req rabbit.Request) error
}

func New(repo spaceRepo, cache cache, saver dbWorker) *Space {
	return &Space{repo: repo, cache: cache, saver: saver}
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

// UpdateNote отправляет запрос на обновление заметки в db-worker
func (s *Space) UpdateNote(ctx context.Context, reqID uuid.UUID, update model.UpdateNote) error {
	req := rabbit.Request{
		ID:   reqID,
		Data: update,
	}

	return s.saver.AddQuery(req)
}

// CheckIfNoteExistsInSpace проверяет, что в пространстве существует такая заметка
func (s *Space) CheckIfNoteExistsInSpace(ctx context.Context, noteID, spaceID uuid.UUID) error {
	return s.repo.CheckIfNoteExistsInSpace(ctx, noteID, spaceID)
}

// IsUserInSpace проверяет, состоит ли пользователь в пространстве
func (s *Space) IsUserInSpace(ctx context.Context, userID int64, spaceID uuid.UUID) error {
	err := s.repo.CheckParticipant(ctx, userID, spaceID)
	if err == nil {
		return nil
	}

	return api_errors.ErrSpaceNotBelongsUser
}
