package model

import "time"

// Space - пространство пользователя. Может быть личным или совместным (указано во флаге personal bool)
type Space struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Created  time.Time `json:"created"`  // указывается в часовом поясе пользователя-создателя
	Creator  int64     `json:"creator"`  // айди пользователя-создателя в телеге
	Personal bool      `json:"personal"` // личное / совместное пространство
}
