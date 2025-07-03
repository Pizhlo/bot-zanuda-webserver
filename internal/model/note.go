package model

import (
	"database/sql"
	"errors"
	"fmt"

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
	ErrUpdateNotTextNote = errors.New("not possible to update a non-text note")
	// ошибка о том, что поле space_id заполнено неправильно
	ErrInvalidSpaceID = errors.New("invalid space id")
	// ошибка о том, что не заполнено поле id
	ErrIDNotFilled = errors.New("field `id` not filled")
)

// тип заметки
type NoteType string

const (
	// текстовая заметка
	TextNoteType NoteType = "text"
	// заметка с фото
	PhotoNoteType NoteType = "photo"
)

type Note struct {
	ID      uuid.UUID     `json:"id"`
	User    *User         `json:"user"`    // кто создал заметку
	Text    string        `json:"text"`    // текст заметки
	Space   *Space        `json:"space"`   // айди пространства, куда сохранить заметку
	Created int64         `json:"created"` // дата создания заметки в часовом поясе пользователя в unix
	Updated sql.NullInt64 `json:"updated"`
	Type    NoteType      `json:"type"` // тип заметки: текстовая, фото, видео, етс
	File    string        `json:"file"` // название файла в Minio (если есть)
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
	ID      uuid.UUID      `json:"id"`
	UserID  int            `json:"user_id"`
	Text    string         `json:"text"`
	SpaceID uuid.UUID      `json:"space_id"`
	Created int64          `json:"created"` // дата создания заметки в часовом поясе пользователя в unix
	Updated sql.NullInt64  `json:"updated"`
	Type    NoteType       `json:"type"`
	File    sql.NullString `json:"file"` // название файла в Minio (если есть)
}

func (s *GetNote) Validate() error {
	if s.UserID == 0 {
		return ErrFieldUserNotFilled
	}

	if s.Text == "" {
		return ErrFieldTextNotFilled
	}

	if s.Created == 0 {
		return ErrFieldCreatedNotFilled
	}

	// можем не валидировать uuid, т.к. если он будет invalid, то структура просто не спарсится
	if s.SpaceID == uuid.Nil {
		return ErrInvalidSpaceID
	}

	if len(s.Type) == 0 {
		return ErrFieldTypeNotFilled
	}

	return nil
}

// структура для ответа на запрос всех типов заметок
type NoteTypeResponse struct {
	Type  NoteType `json:"type"`
	Count int      `json:"count"`
}

// запрос на поиск заметок по тексту в пространстве
type SearchNoteByTextRequest struct {
	SpaceID uuid.UUID `json:"space_id"`
	Text    string    `json:"text"`
	Type    NoteType  `json:"type"` // тип заметок, для которого осуществлять поиск
}
