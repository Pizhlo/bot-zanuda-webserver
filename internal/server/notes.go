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

//	@Summary		Запрос на создание заметки
//	@Description	Запрос на создание заметки с текстом для определенного пользователя
//	@Param			request	body	model.CreateNoteRequest	true	"создать заметку:\nуказать айди пользователя,\nайди его личного / совместного пространства,\nтекст заметки,\nдата создания в часовом поясе пользователя в unix"
//	@Success		201
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/notes/create [post]
//
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

	// пространство личное - проверяем, принадлежит ли оно тому же пользователю, что и отправил запрос
	if req.Space.Personal {
		space, err := s.space.GetSpaceByID(c.Request().Context(), req.Space.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		if space.Creator != req.UserID {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": "space is personal and belongs to another user"})
		}
	}

	err = s.note.Create(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, note.ErrUnknownUser) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		if errors.Is(err, note.ErrSpaceNotExists) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusCreated)
}
