package model

import "errors"

type User struct {
	ID            int    `json:"id"`
	TgID          int64  `json:"tg_id"`
	Username      string `json:"username"`
	PersonalSpace *Space `json:"personal_space"`
	Timezone      string `json:"timezone"`
}

var (
	// не заполнено поле tg_id
	ErrTgIDNotFilled = errors.New("field `TgID` not filled")
	// не заполнено поле username
	ErrUsernameNotFilled = errors.New("field `username` not filled")
	// не заполнено поле timezone
	ErrTimezoneNotFilled = errors.New("field `timezone` not filled")
)

func (s *User) Validate() error {
	if s.ID == 0 {
		return ErrFieldIDNotFilled
	}

	if s.TgID == 0 {
		return ErrTgIDNotFilled
	}

	if len(s.Username) == 0 {
		return ErrUsernameNotFilled
	}

	if s.PersonalSpace == nil {
		return ErrSpaceIsNil
	}

	if err := s.PersonalSpace.Validate(); err != nil {
		return err
	}

	if len(s.Timezone) == 0 {
		return ErrTimezoneNotFilled
	}

	return nil
}
