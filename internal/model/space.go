package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Space - пространство пользователя. Может быть личным или совместным (указано во флаге personal bool)
type Space struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Created  time.Time `json:"created"`  // указывается в часовом поясе пользователя-создателя
	Creator  int64     `json:"creator"`  // айди пользователя-создателя в телеге
	Personal bool      `json:"personal"` // личное / совместное пространство
}

var (
	ErrFieldIDNotFilled = errors.New("field `id` not filled")
	// не заполнено поле Name
	ErrFieldNameNotFilled = errors.New("field `name` not filled")
	// не заполнено поле Creator
	ErrFieldCreatorNotFilled = errors.New("field `creator` not filled")
)

func (s *Space) Validate() error {
	if s.ID == uuid.Nil {
		return ErrFieldIDNotFilled
	}

	if len(s.Name) == 0 {
		return ErrFieldNameNotFilled
	}

	if s.Created.IsZero() {
		return ErrFieldCreatedNotFilled
	}

	if s.Creator == 0 {
		return ErrFieldCreatorNotFilled
	}

	return nil
}
