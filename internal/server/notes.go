package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"webserver/internal/model"
	"webserver/internal/service/storage/postgres/note"
	"webserver/internal/service/storage/postgres/space"

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

//	@Summary		Запрос на получение всех заметок
//	@Description	Запрос на получение всех заметок из личного пространства пользователя
//	@Success		200 {object}    []model.Note
//	@Success		204
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			spaces/:id/notes [get]
//
// ручка для получения всех заметок пользователя из его личного пространства
func (s *server) notesBySpaceID(c echo.Context) error {
	spaceIDStr := c.Param("id")

	spaceID, err := strconv.Atoi(spaceIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid space id parameter: %+v", err)})
	}

	// предоставлять ли полную инф-ю о пользователе, который создал заметку
	fullUserParam := c.QueryParam("full_user")

	var fullUser bool // по умолчанию false

	if len(fullUserParam) > 0 {
		var err error
		fullUser, err = strconv.ParseBool(fullUserParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid full user parameter: %+v", err)})
		}
	}

	// получение заметок в полном режиме
	if fullUser {
		notes, err := s.space.GetAllbySpaceIDFull(c.Request().Context(), int64(spaceID))
		if err != nil {
			// у пользователя нет заметок - отдаем 204
			if errors.Is(err, space.ErrNoNotesFoundBySpaceID) {
				return c.NoContent(http.StatusNoContent)
			}

			// пространство не существует - отдаем 404
			if errors.Is(err, space.ErrSpaceNotExists) {
				return c.NoContent(http.StatusNotFound)
			}

			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, notes)
	}

	// получение заметок в кратком режиме
	notes, err := s.space.GetAllBySpaceID(c.Request().Context(), int64(spaceID))
	if err != nil {
		// у пользователя нет заметок - отдаем 204
		if errors.Is(err, space.ErrNoNotesFoundBySpaceID) {
			return c.NoContent(http.StatusNoContent)
		}

		// пространство не существует - отдаем 404
		if errors.Is(err, space.ErrSpaceNotExists) {
			return c.NoContent(http.StatusNotFound)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, notes)
}
