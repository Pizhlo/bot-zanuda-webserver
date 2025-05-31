package space

import (
	"context"
	"errors"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"

	"github.com/google/uuid"
)

type Space struct {
	repo   repo
	cache  spaceCache
	worker dbWorker // создание / обновление записей
}

//go:generate mockgen -source ./space.go -destination=../../../mocks/space_srv.go -package=mocks
type repo interface {
	noteRepo
	spaceRepo
}

//go:generate mockgen -source ./space.go -destination=../../../mocks/space_srv.go -package=mocks
type spaceRepo interface {
	GetSpaceByID(ctx context.Context, id uuid.UUID) (model.Space, error)
	// CheckParticipant проверяет, является ли пользователь участником пространства
	CheckParticipant(ctx context.Context, userID int64, spaceID uuid.UUID) error
}

//go:generate mockgen -source ./space.go -destination=../../../mocks/space_srv.go -package=mocks
type noteRepo interface {
	// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает в полном виде.
	GetAllNotesBySpaceIDFull(ctx context.Context, spaceID uuid.UUID) ([]model.Note, error)
	// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает кратко (только userID)
	GetAllNotesBySpaceID(ctx context.Context, spaceID uuid.UUID) ([]model.GetNote, error)
	// GetNoteByID возвращает заметку по айди, либо ошибку о том, что такой заметки не существует
	GetNoteByID(ctx context.Context, noteID uuid.UUID) (model.GetNote, error)
	// GetNotesTypes возвращает все типы заметок в пространстве и их количество (3 текстовых, 2 фото, и т.п.)
	GetNotesTypes(ctx context.Context, spaceID uuid.UUID) ([]model.NoteTypeResponse, error)
	// GetNotesByType возвращает все заметки указанного типа из пространства
	GetNotesByType(ctx context.Context, spaceID uuid.UUID, noteType model.NoteType) ([]model.GetNote, error)
	SearchNoteByText(ctx context.Context, req model.SearchNoteByTextRequest) ([]model.GetNote, error)
}

//go:generate mockgen -source ./service.go -destination=../../../mocks/space_srv.go -package=mocks
type spaceCache interface {
	GetSpaceByID(ctx context.Context, id uuid.UUID) (model.Space, error)
}

// dbWorker работает на создание / обновление записей
type dbWorker interface {
	noteEditor
	spaceEditor
}

type spaceEditor interface {
	CreateSpace(ctx context.Context, req rabbit.CreateSpaceRequest) error
}

type noteEditor interface {
	CreateNote(ctx context.Context, req rabbit.Model) error
	UpdateNote(ctx context.Context, req rabbit.Model) error
	DeleteNote(ctx context.Context, req rabbit.Model) error
	DeleteAllNotes(ctx context.Context, req rabbit.Model) error
}

func New(repo repo, cache spaceCache, worker dbWorker) (*Space, error) {
	if repo == nil {
		return nil, errors.New("repo is nil")
	}

	if cache == nil {
		return nil, errors.New("cache is nil")
	}

	if worker == nil {
		return nil, errors.New("worker is nil")
	}

	return &Space{repo: repo, cache: cache, worker: worker}, nil
}
