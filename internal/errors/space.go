package errors

import "errors"

var (
	// ошибка о том, что пользователь не может добавить запись в это пространство: оно личное и не принадлежит ему
	ErrSpaceNotBelongsUser = errors.New("space not belongs to user")
	// ошибка о том, что в пространстве нет заметок
	ErrNoNotesFoundBySpaceID = errors.New("space does not have any notes")
	// ошибка о том, что пространство не существует
	ErrSpaceNotExists = errors.New("space does not exist")
)
