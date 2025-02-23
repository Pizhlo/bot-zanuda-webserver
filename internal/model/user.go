package model

type User struct {
	ID            int    `json:"id"`
	TgID          int64  `json:"tg_id"`
	Username      string `json:"username"`
	PersonalSpace Space  `json:"personal_space"`
}

func (s *User) Validate() error {
	return nil
}
