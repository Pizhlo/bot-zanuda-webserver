package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	// ошибка о том, что не заполнено поле user_id
	ErrFieldUserNotFilled = errors.New("field `user_id` not filled")
	// ошибка о том, что не заполнено поле text
	ErrFieldTextNotFilled = errors.New("field `text` not filled")
	// ошибка о том, что не заполнено поле created
	ErrFieldCreatedNotFilled = errors.New("field `created` not filled")
	// ошибка о том, что не заполнено поле type
	ErrFieldTypeNotFilled = errors.New("field `type` not filled")
	// ошибка о том, что произошла попытка обновить не текстовую заметку
	ErrUpdateNotTextNote = errors.New("field `type` is not text. Unable to update note")
	// ошибка о том, что поле space_id заполнено неправильно
	ErrInvalidSpaceID = errors.New("invalid space id")
	// ошибка о том, что не заполнено поле id у заметки
	ErrNoteIdNotFilled = errors.New("field `id` not filled")
)

// тип заметки
type NoteType string

const (
	// текстовая заметка
	TextNoteType NoteType = "text"
	// заметка с фото
	PhotoNoteType NoteType = "photo"
)

//	{
//	  "user_id": 12345678,
//	  "text": "new note",
//	  “space_id” :1,
//	  "created": 1739264640
//	}
//
// Запрос на создание заметки
type CreateNoteRequest struct {
	UserID  int64     `json:"user_id"`  // кто создал заметку
	SpaceID uuid.UUID `json:"space_id"` // айди пространства, куда сохранить заметку
	Text    string    `json:"text"`     // текст заметки
	Created int64     `json:"created"`  // дата создания заметки в часовом поясе пользователя в unix
	Type    NoteType  `json:"type"`     // тип заметки: текстовая, фото, видео, етс
}

func (s *CreateNoteRequest) Validate() error {
	// проверяем не все поля, т.к. не все поля заполнены из запроса
	if s.UserID == 0 {
		return ErrFieldUserNotFilled
	}

	if s.Text == "" {
		return ErrFieldTextNotFilled
	}

	if err := uuid.Validate(s.SpaceID.String()); err != nil || s.SpaceID == uuid.Nil {
		return ErrInvalidSpaceID
	}

	if len(s.Type) == 0 {
		return ErrFieldTypeNotFilled
	}

	return nil
}

type Note struct {
	ID       uuid.UUID    `json:"id"`
	User     *User        `json:"user"`    // кто создал заметку
	Text     string       `json:"text"`    // текст заметки
	Space    *Space       `json:"space"`   // айди пространства, куда сохранить заметку
	Created  time.Time    `json:"created"` // дата создания заметки в часовом поясе пользователя в unix
	LastEdit sql.NullTime `json:"last_edit"`
	Type     NoteType     `json:"type"` // тип заметки: текстовая, фото, видео, етс
}

var ErrSpaceIsNil = fmt.Errorf("field `Space` is nil")

func (s *Note) Validate() error {
	if s.User == nil {
		return ErrFieldUserNotFilled
	}

	if err := s.User.Validate(); err != nil {
		return err
	}

	if s.Text == "" {
		return ErrFieldTextNotFilled
	}

	if s.Space == nil {
		return ErrSpaceIsNil
	}

	if err := s.Space.Validate(); err != nil {
		return err
	}

	if len(s.Type) == 0 {
		return ErrFieldTypeNotFilled
	}

	return nil
}

// структура для ответа на запрос всех заметок пространства в кратком режиме.
// У этой структуры поля пользователь и пространство заменены на айди
type GetNote struct {
	ID       uuid.UUID    `json:"id"`
	UserID   int          `json:"user_id"`
	Text     string       `json:"text"`
	SpaceID  uuid.UUID    `json:"space_id"`
	Created  time.Time    `json:"created"` // дата создания заметки в часовом поясе пользователя в unix
	LastEdit sql.NullTime `json:"last_edit"`
	Type     NoteType     `json:"type"`
}

func (s *GetNote) Validate() error {
	if s.UserID == 0 {
		return ErrFieldUserNotFilled
	}

	if s.Text == "" {
		return ErrFieldTextNotFilled
	}

	if s.Created.IsZero() {
		return ErrFieldCreatedNotFilled
	}

	if err := uuid.Validate(s.SpaceID.String()); err != nil {
		return ErrInvalidSpaceID
	}

	if len(s.Type) == 0 {
		return ErrFieldTypeNotFilled
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
type UpdateNote struct {
	SpaceID uuid.UUID `json:"space_id"`
	UserID  int64     `json:"user_id"`
	ID      uuid.UUID `json:"id"`   // айди заметки
	Text    string    `json:"text"` // новый текст
	Type    NoteType  `json:"type"`
}

func (s *UpdateNote) Validate() error {
	if s.UserID == 0 {
		return ErrFieldUserNotFilled
	}

	if s.Text == "" {
		return ErrFieldTextNotFilled
	}

	if err := uuid.Validate(s.SpaceID.String()); err != nil {
		return ErrInvalidSpaceID
	}

	if err := uuid.Validate(s.ID.String()); err != nil {
		return ErrInvalidSpaceID
	}

	if len(s.Type) == 0 {
		return ErrFieldTypeNotFilled
	}

	if s.Type != TextNoteType {
		return ErrUpdateNotTextNote
	}

	return nil
}
