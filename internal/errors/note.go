package errors

import "errors"

var (
	// ошибка о том, что в пространстве нет такой заметки
	ErrNoteNotBelongsSpace = errors.New("note does not belong space")
	// заметка не найдена
	ErrNoteNotFound = errors.New("note not found")
	// ошибка о том, что не найдены заметки указанного типа
	ErrNoNotesFoundByType = errors.New("no notes found by this type")
	// ошибка о том, что заметки по тексту не найдены
	ErrNoNotesFoundByText = errors.New("no notes found by text")
)
