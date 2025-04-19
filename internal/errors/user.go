package errors

import "errors"

var (
	// ошибка о том, что пользователя не существует в БД
	ErrUnknownUser = errors.New("unknown user")
)
