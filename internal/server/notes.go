package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"webserver/internal/model"
	"webserver/internal/service/storage/postgres/note"

	"github.com/labstack/echo/v4"
)

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

	err = s.note.Create(c.Request().Context(), req)
	if err != nil {
		// неизвестный пользователь
		if errors.Is(err, note.ErrUnknownUser) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		// пространства не существует
		if errors.Is(err, note.ErrSpaceNotExists) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		// пространство личное и принадлежит другому пользователю
		if errors.Is(err, note.ErrSpaceNotBelongsUser) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusCreated)
}

func (s *server) notesByUserID(c echo.Context) error {
	userIDStr := c.Param("id")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid user id param: %+v", err)})
	}

	notes, err := s.note.GetAllbyUserID(c.Request().Context(), int64(userID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// добавить проверку на то что нет заметок - 204

	return c.JSON(http.StatusOK, notes)
}
