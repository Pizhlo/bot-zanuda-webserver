package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
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
	UserID  int64  `json:"user_id"`  // кто создал заметку
	SpaceID int    `json:"space_id"` // айди пространства, куда сохранить заметку
	Text    string `json:"text"`     // текст заметки
	Created int64  `json:"created"`  // дата создания заметки в часовом поясе пользователя в unix
}

func (s *CreateNoteRequest) Validate() error {
	// проверяем не все поля, т.к. не все поля заполнены из запроса
	if s.UserID == 0 {
		return fmt.Errorf("field `user_id` not filled")
	}

	if s.Text == "" {
		return fmt.Errorf("field `text` not filled")
	}

	if s.Created == 0 {
		return fmt.Errorf("field `created` not filled")
	}

	if s.SpaceID == 0 {
		return fmt.Errorf("field `space_id` not filled")
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
}

func (s *Note) Validate() error {
	if s.User == nil {
		return fmt.Errorf("field `User` is nil")
	}

	if err := s.User.Validate(); err != nil {
		return err
	}

	if s.Text == "" {
		return fmt.Errorf("field `text` not filled")
	}

	if s.Created.IsZero() {
		return fmt.Errorf("field `created` not filled")
	}

	if s.Space == nil {
		return fmt.Errorf("field `Space` is nil")
	}

	if err := s.Space.Validate(); err != nil {
		return err
	}

	return nil
}
