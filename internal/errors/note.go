package errors

import "errors"

var (
	// ошибка о том, что в пространстве нет такой заметки
	ErrNoteNotBelongsSpace = errors.New("note does not belong space")
)
