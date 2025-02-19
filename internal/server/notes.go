package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"webserver/internal/model"
	"webserver/internal/service/storage/postgres/note"

	"github.com/labstack/echo/v4"
)

func (s *server) notesByUserID(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

// ручка для создания заметки
func (s *server) createNote(c echo.Context) error {
	var req model.CreateNoteRequest

	err := json.NewDecoder(c.Request().Body).Decode(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	// проверяем поля на валидность
	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	err = s.note.Create(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, note.ErrUnknownUser) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		if errors.Is(err, note.ErrSpaceNotExists) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusCreated)
}
