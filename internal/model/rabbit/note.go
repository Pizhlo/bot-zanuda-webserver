package rabbit

import (
	"webserver/internal/model"

	"github.com/google/uuid"
)

// интерфейс для всех моделей rabbit
//
//go:generate mockgen -source ./note.go -destination=./mocks/rabbit.go -package=mocks
type Model interface {
	Validate() error
	GetID() uuid.UUID
}

//	{
//	  "user_id": 12345678,
//	  "text": "new note",
//	  “space_id” :1,
//	  "created": 1739264640
//	}
//
// Запрос на создание заметки
type CreateNoteRequest struct {
	ID        uuid.UUID      `json:"request_id"` // айди запроса
	UserID    int64          `json:"user_id"`    // кто создал заметку
	SpaceID   uuid.UUID      `json:"space_id"`   // айди пространства, куда сохранить заметку
	Text      string         `json:"text"`       // текст заметки
	Type      model.NoteType `json:"type"`       // тип заметки: текстовая, фото, видео, етс
	File      string         `json:"file"`       // название файла в Minio (если есть)
	Operation Operation      `json:"operation"`  // какое действие сделать: создать, удалить, редактировать
	Created   int64          `json:"created"`    // дата обращения в Unix в UTC
}

func (s *CreateNoteRequest) GetID() uuid.UUID {
	return s.ID
}

func (s *CreateNoteRequest) Validate() error {
	if s.ID == uuid.Nil {
		return model.ErrIDNotFilled
	}

	if s.UserID == 0 {
		return model.ErrFieldUserNotFilled
	}

	if s.Text == "" {
		return model.ErrFieldTextNotFilled
	}

	// можем не валидировать uuid, т.к. если он будет invalid, то структура просто не спарсится
	if s.SpaceID == uuid.Nil {
		return model.ErrInvalidSpaceID
	}

	if len(s.Type) == 0 {
		return model.ErrFieldTypeNotFilled
	}

	if s.Created == 0 {
		return model.ErrFieldCreatedNotFilled
	}

	if s.Operation != CreateOp {
		return ErrInvalidOperation
	}

	return nil
}

// структура для запроса на обновление заметки.
// обновляются текст и last_update
//
//	{
//		"space_id": "ed3a5b3a-b81e-4cad-acea-178e230a9b93”,
//		"user_id": 12354,
//		"id": “ed3a5b3a-b81e-4cad-acea-178e230a9b93”,
//		“text”: “new note text"
//	  }
type UpdateNoteRequest struct {
	ID        uuid.UUID `json:"request_id"` // айди запроса, генерируется в процессе обработки
	SpaceID   uuid.UUID `json:"space_id"`
	UserID    int64     `json:"user_id"`
	NoteID    uuid.UUID `json:"note_id"`   // айди заметки
	Text      string    `json:"text"`      // новый текст
	File      string    `json:"file"`      // название файла в Minio (если есть)
	Operation Operation `json:"operation"` // какое действие сделать: создать, удалить, редактировать
	Created   int64     `json:"created"`   // дата обращения в Unix в UTC
}

func (s *UpdateNoteRequest) GetID() uuid.UUID {
	return s.ID
}

func (s *UpdateNoteRequest) Validate() error {
	if s.ID == uuid.Nil {
		return model.ErrIDNotFilled
	}

	if s.UserID == 0 {
		return model.ErrFieldUserNotFilled
	}

	if s.Text == "" {
		return model.ErrFieldTextNotFilled
	}

	// можем не валидировать uuid, т.к. если он будет invalid, то структура просто не спарсится
	if s.SpaceID == uuid.Nil {
		return model.ErrInvalidSpaceID
	}

	if s.Created == 0 {
		return model.ErrFieldCreatedNotFilled
	}

	if s.Operation != UpdateOp {
		return ErrInvalidOperation
	}

	return nil
}

type DeleteNoteRequest struct {
	ID        uuid.UUID `json:"request_id"` // айди запроса, генерируется в процессе обработки
	SpaceID   uuid.UUID `json:"space_id"`
	NoteID    uuid.UUID `json:"note_id"`
	Operation Operation `json:"operation"` // какое действие сделать: создать, удалить, редактировать
	Created   int64     `json:"created"`   // дата обращения в Unix в UTC
}

func (s *DeleteNoteRequest) GetID() uuid.UUID {
	return s.ID
}

func (s *DeleteNoteRequest) Validate() error {
	if s.ID == uuid.Nil {
		return model.ErrFieldIDNotFilled
	}

	if s.SpaceID == uuid.Nil {
		return model.ErrInvalidSpaceID
	}

	if s.NoteID == uuid.Nil {
		return model.ErrIDNotFilled
	}

	if s.Created == 0 {
		return model.ErrFieldCreatedNotFilled
	}

	if s.Operation != DeleteOp {
		return ErrInvalidOperation
	}

	return nil
}

type DeleteAllNotesRequest struct {
	ID        uuid.UUID `json:"request_id"` // айди запроса, генерируется в процессе обработки
	SpaceID   uuid.UUID `json:"space_id"`
	Operation Operation `json:"operation"` // какое действие сделать: создать, удалить, редактировать
	Created   int64     `json:"created"`   // дата обращения в Unix в UTC
}

func (s *DeleteAllNotesRequest) GetID() uuid.UUID {
	return s.ID
}

func (s *DeleteAllNotesRequest) Validate() error {
	if s.ID == uuid.Nil {
		return model.ErrFieldIDNotFilled
	}

	if s.SpaceID == uuid.Nil {
		return model.ErrInvalidSpaceID
	}

	if s.Created == 0 {
		return model.ErrFieldCreatedNotFilled
	}

	if s.Operation != DeleteAllOp {
		return ErrInvalidOperation
	}

	return nil
}
