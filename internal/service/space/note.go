package space

import (
	"context"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"

	"github.com/google/uuid"
)

// Create создает новую заметку в личном пространстве пользователя
func (s *Space) CreateNote(ctx context.Context, note rabbit.CreateNoteRequest) error {
	return s.worker.CreateNote(ctx, &note)
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
func (s *Space) UpdateNote(ctx context.Context, update rabbit.UpdateNoteRequest) error {
	return s.worker.UpdateNote(ctx, &update)
}

// GetNoteByID возвращает заметку по айди, либо ошибку о том, что такой заметки не существует
func (s *Space) GetNoteByID(ctx context.Context, noteID uuid.UUID) (model.GetNote, error) {
	return s.repo.GetNoteByID(ctx, noteID)
}

// GetNotesTypes возвращает все типы заметок в пространстве и их количество (3 текстовых, 2 фото, и т.п.)
func (s *Space) GetNotesTypes(ctx context.Context, spaceID uuid.UUID) ([]model.NoteTypeResponse, error) {
	return s.repo.GetNotesTypes(ctx, spaceID)
}

// GetNotesByType возвращает все заметки указанного типа из пространства
func (s *Space) GetNotesByType(ctx context.Context, spaceID uuid.UUID, noteType model.NoteType) ([]model.GetNote, error) {
	return s.repo.GetNotesByType(ctx, spaceID, noteType)
}

func (s *Space) SearchNoteByText(ctx context.Context, req model.SearchNoteByTextRequest) ([]model.GetNote, error) {
	if len(req.Type) == 0 { // по умолчанию, если не указано, ищем среди текстовых
		req.Type = model.TextNoteType
	}

	return s.repo.SearchNoteByText(ctx, req)
}

func (s *Space) DeleteNote(ctx context.Context, req rabbit.DeleteNoteRequest) error {
	return s.worker.DeleteNote(ctx, &req)
}

func (s *Space) DeleteAllNotes(ctx context.Context, req rabbit.DeleteAllNotesRequest) error {
	return s.worker.DeleteAllNotes(ctx, &req)
}
