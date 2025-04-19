package space

import (
	"context"
	"errors"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"

	"github.com/google/uuid"
)

type Space struct {
	repo  spaceRepo
	cache spaceCache
	saver dbWorker // создание / обновление записей
}

//go:generate mockgen -source ./space.go -destination=../../../mocks/space_srv.go -package=mocks
type spaceRepo interface {
	GetSpaceByID(ctx context.Context, id uuid.UUID) (model.Space, error)
	// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает в полном виде.
	GetAllNotesBySpaceIDFull(ctx context.Context, spaceID uuid.UUID) ([]model.Note, error)
	// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает кратко (только userID)
	GetAllNotesBySpaceID(ctx context.Context, spaceID uuid.UUID) ([]model.GetNote, error)
	// CheckParticipant проверяет, является ли пользователь участником пространства
	CheckParticipant(ctx context.Context, userID int64, spaceID uuid.UUID) error
	// GetNoteByID возвращает заметку по айди, либо ошибку о том, что такой заметки не существует
	GetNoteByID(ctx context.Context, noteID uuid.UUID) (model.GetNote, error)
	// GetNotesTypes возвращает все типы заметок в пространстве и их количество (3 текстовых, 2 фото, и т.п.)
	GetNotesTypes(ctx context.Context, spaceID uuid.UUID) ([]model.NoteTypeResponse, error)
	// GetNotesByType возвращает все заметки указанного типа из пространства
	GetNotesByType(ctx context.Context, spaceID uuid.UUID, noteType model.NoteType) ([]model.GetNote, error)
}

type spaceCache interface {
	GetSpaceByID(ctx context.Context, id uuid.UUID) (model.Space, error)
}

// dbWorker работает на создание / обновление записей
type dbWorker interface {
	CreateNote(req model.CreateNoteRequest) error
	UpdateNote(req model.UpdateNoteRequest) error
}

func New(repo spaceRepo, cache spaceCache, saver dbWorker) *Space {
	return &Space{repo: repo, cache: cache, saver: saver}
}

func (s *Space) GetSpaceByID(ctx context.Context, id uuid.UUID) (model.Space, error) {
	space, err := s.cache.GetSpaceByID(ctx, id)
	if err != nil {
		if !errors.Is(err, api_errors.ErrSpaceNotExists) {
			return model.Space{}, err
		}
	}

	if err == nil {
		return space, nil
	}

	return s.repo.GetSpaceByID(ctx, id)
}

// Create создает новую заметку в личном пространстве пользователя
func (s *Space) CreateNote(ctx context.Context, note model.CreateNoteRequest) error {
	return s.saver.CreateNote(note)
}

// GetAllNotesBySpaceIDFull возвращает все заметки пространства.
// Информацию о пользователе возвращает в полном виде.
func (s *Space) GetAllNotesBySpaceIDFull(ctx context.Context, spaceID uuid.UUID) ([]model.Note, error) {
	return s.repo.GetAllNotesBySpaceIDFull(ctx, spaceID)
}

// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает кратко (только userID)
func (s *Space) GetAllNotesBySpaceID(ctx context.Context, spaceID uuid.UUID) ([]model.GetNote, error) {
	return s.repo.GetAllNotesBySpaceID(ctx, spaceID)
}

// UpdateNote отправляет запрос на обновление заметки в db-worker
func (s *Space) UpdateNote(ctx context.Context, update model.UpdateNoteRequest) error {
	return s.saver.UpdateNote(update)
}

// GetNoteByID возвращает заметку по айди, либо ошибку о том, что такой заметки не существует
func (s *Space) GetNoteByID(ctx context.Context, noteID uuid.UUID) (model.GetNote, error) {
	return s.repo.GetNoteByID(ctx, noteID)
}

// IsUserInSpace проверяет, состоит ли пользователь в пространстве
func (s *Space) IsUserInSpace(ctx context.Context, userID int64, spaceID uuid.UUID) error {
	return s.repo.CheckParticipant(ctx, userID, spaceID)
}

// GetNotesTypes возвращает все типы заметок в пространстве и их количество (3 текстовых, 2 фото, и т.п.)
func (s *Space) GetNotesTypes(ctx context.Context, spaceID uuid.UUID) ([]model.NoteTypeResponse, error) {
	return s.repo.GetNotesTypes(ctx, spaceID)
}

// GetNotesByType возвращает все заметки указанного типа из пространства
func (s *Space) GetNotesByType(ctx context.Context, spaceID uuid.UUID, noteType model.NoteType) ([]model.GetNote, error) {
	return s.repo.GetNotesByType(ctx, spaceID, noteType)
}
