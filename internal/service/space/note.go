package space

import (
	"context"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"

	"github.com/google/uuid"
)

// Create создает новую заметку в личном пространстве пользователя
func (s *Service) CreateNote(ctx context.Context, note rabbit.CreateNoteRequest) error {
	s.logger.WithField("request_id", note.ID).Debug("creating note")
	return s.worker.CreateNote(ctx, &note)
}

// GetAllNotesBySpaceIDFull возвращает все заметки пространства.
// Информацию о пользователе возвращает в полном виде.
func (s *Service) GetAllNotesBySpaceIDFull(ctx context.Context, spaceID uuid.UUID) ([]model.Note, error) {
	s.logger.WithField("space_id", spaceID).Debug("getting all notes by space id full")
	return s.repo.GetAllNotesBySpaceIDFull(ctx, spaceID)
}

// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает кратко (только userID)
func (s *Service) GetAllNotesBySpaceID(ctx context.Context, spaceID uuid.UUID) ([]model.GetNote, error) {
	s.logger.WithField("space_id", spaceID).Debug("getting all notes by space id")
	return s.repo.GetAllNotesBySpaceID(ctx, spaceID)
}

// UpdateNote отправляет запрос на обновление заметки в db-worker
func (s *Service) UpdateNote(ctx context.Context, update rabbit.UpdateNoteRequest) error {
	s.logger.WithField("request_id", update.NoteID).Debug("updating note")
	return s.worker.UpdateNote(ctx, &update)
}

// GetNoteByID возвращает заметку по айди, либо ошибку о том, что такой заметки не существует
func (s *Service) GetNoteByID(ctx context.Context, noteID uuid.UUID) (model.GetNote, error) {
	s.logger.WithField("note_id", noteID).Debug("getting note by id")
	return s.repo.GetNoteByID(ctx, noteID)
}

// GetNotesTypes возвращает все типы заметок в пространстве и их количество (3 текстовых, 2 фото, и т.п.)
func (s *Service) GetNotesTypes(ctx context.Context, spaceID uuid.UUID) ([]model.NoteTypeResponse, error) {
	s.logger.WithField("space_id", spaceID).Debug("getting notes types")
	return s.repo.GetNotesTypes(ctx, spaceID)
}

// GetNotesByType возвращает все заметки указанного типа из пространства
func (s *Service) GetNotesByType(ctx context.Context, spaceID uuid.UUID, noteType model.NoteType) ([]model.GetNote, error) {
	s.logger.WithField("space_id", spaceID).WithField("noteType", noteType).Debug("getting notes by type")
	return s.repo.GetNotesByType(ctx, spaceID, noteType)
}

func (s *Service) SearchNoteByText(ctx context.Context, req model.SearchNoteByTextRequest) ([]model.GetNote, error) {
	s.logger.WithField("space_id", req.SpaceID).Debug("searching note by text")

	if len(req.Type) == 0 { // по умолчанию, если не указано, ищем среди текстовых
		req.Type = model.TextNoteType
	}

	return s.repo.SearchNoteByText(ctx, req)
}

func (s *Service) DeleteNote(ctx context.Context, req rabbit.DeleteNoteRequest) error {
	s.logger.WithField("request_id", req.ID).Debug("deleting note")
	return s.worker.DeleteNote(ctx, &req)
}

func (s *Service) DeleteAllNotes(ctx context.Context, req rabbit.DeleteAllNotesRequest) error {
	s.logger.WithField("request_id", req.ID).Debug("deleting all notes")
	return s.worker.DeleteAllNotes(ctx, &req)
}
