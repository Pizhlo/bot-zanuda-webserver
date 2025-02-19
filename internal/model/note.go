package model

import "fmt"

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
	Text    string `json:"text"`     // текст заметки
	SpaceID int64  `json:"space_id"` // айди пространства (личного или совместного), куда сохранить заметку
	Created int64  `json:"created"`  // дата создания заметки в часовом поясе пользователя в unix
}

func (s *CreateNoteRequest) Validate() error {
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
