package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

//	@Summary		Запрос на создание заметки
//	@Description	Запрос на создание заметки с текстом. Создается в указанном пространстве
//	@Param			request	body	model.CreateNoteRequest	true	"создать заметку:\nуказать айди пользователя,\nайди его личного / совместного пространства,\nтекст заметки,\nдата создания в часовом поясе пользователя в unix"
//	@Success		201
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/spaces/notes/create [post]
//
// ручка для создания заметки
func (s *server) createNote(c echo.Context) error {
	var req model.CreateNoteRequest

	err := json.NewDecoder(c.Request().Body).Decode(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	reqID := uuid.New()

	// TODO: переделать на rabbitMQ
	err = s.space.CreateNote(c.Request().Context(), reqID, req)
	if err != nil {
		// ошибки запроса
		errs := []error{
			model.ErrSpaceIdNotFilled, model.ErrFieldCreatedNotFilled,
			model.ErrFieldTextNotFilled, model.ErrNoteIdNotFilled,
			model.ErrFieldUserNotFilled, api_errors.ErrUnknownUser,
			api_errors.ErrSpaceNotExists, api_errors.ErrSpaceNotBelongsUser,
		}

		if errorsIn(err, errs) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		// внутренняя ошибка
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusAccepted, map[string]string{"request_id": reqID.String()})
}

func errorsIn(target error, errs []error) bool {
	for _, err := range errs {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}

//		@Summary		Запрос на получение всех заметок
//		@Description	Запрос на получение всех заметок из личного пространства пользователя
//	 @Param        id   path      int  true  "ID пространства"
//		@Success		200 {object}    []model.Note
//		@Success		200 {object}    []model.GetNote
//		@Success		204                               "В пространстве отсутствют заметки"
//		@Failure		400	{object}	map[string]string "Невалидный запрос"
//		@Failure		404                               "Пространства не существует"
//		@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//		@Router			/spaces/{id}/notes [get]
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
			if errors.Is(err, api_errors.ErrNoNotesFoundBySpaceID) {
				return c.NoContent(http.StatusNoContent)
			}

			// пространство не существует - отдаем 404
			if errors.Is(err, api_errors.ErrSpaceNotExists) {
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
		if errors.Is(err, api_errors.ErrNoNotesFoundBySpaceID) {
			return c.NoContent(http.StatusNoContent)
		}

		// пространство не существует - отдаем 404
		if errors.Is(err, api_errors.ErrSpaceNotExists) {
			return c.NoContent(http.StatusNotFound)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, notes)
}

func (s *server) updateNote(c echo.Context) error {
	var req model.UpdateNote

	err := json.NewDecoder(c.Request().Body).Decode(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	// валидируем данные
	if err := req.Validate(); err != nil {
		// ошибки запроса
		errs := []error{
			model.ErrSpaceIdNotFilled, model.ErrFieldCreatedNotFilled,
			model.ErrFieldTextNotFilled, model.ErrNoteIdNotFilled,
			model.ErrFieldUserNotFilled, api_errors.ErrUnknownUser,
		}

		if errorsIn(err, errs) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}
	}

	// после валидации - проверяем, что пользователь существует
	if err := s.user.CheckUser(c.Request().Context(), req.UserID); err != nil {
		if errors.Is(err, api_errors.ErrUnknownUser) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// проверяем, что пространство существует
	if _, err := s.space.GetSpaceByID(c.Request().Context(), req.SpaceID); err != nil {
		if errors.Is(err, api_errors.ErrSpaceNotExists) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// проверяем, что пользователь состоит в пространстве (сюда потом еще добавится проверка на права)
	if err := s.space.IsUserInSpace(c.Request().Context(), req.UserID, req.SpaceID); err != nil {
		if errors.Is(err, api_errors.ErrSpaceNotBelongsUser) || errors.Is(err, api_errors.ErrSpaceNotExists) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// проверяем, что в пространстве есть заметка с таким айди
	if err := s.space.CheckIfNoteExistsInSpace(c.Request().Context(), req.ID, req.SpaceID); err != nil {
		if errors.Is(err, api_errors.ErrNoteNotBelongsSpace) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	reqID := uuid.New()

	// TODO: переделать на db worker
	if err := s.space.UpdateNote(c.Request().Context(), reqID, req); err != nil {
		// внутренняя ошибка
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// запрос принят в обработку
	// TODO: добавить возврат requestID
	return c.JSON(http.StatusAccepted, map[string]string{"request_id": reqID.String()})
}
