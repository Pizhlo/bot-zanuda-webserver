package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

//	@Summary		Запрос на создание заметки
//	@Description	Запрос на создание заметки с текстом. Создается в указанном пространстве
//	@Param			request	body	model.CreateNoteRequest	true	"создать заметку:\nуказать айди пользователя,\nайди его личного / совместного пространства,\nтекст заметки\nтип заметки: текстовый, фото, видео, и т.п."
//	@Success		202 {object}    string             айди запроса для отслеживания
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/spaces/notes/create [post]
//
// ручка для создания заметки
func (s *server) createNote(c echo.Context) error {
	var req model.CreateNoteRequest

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	if err := req.Validate(); err != nil {
		// ошибки запроса
		errs := []error{
			model.ErrInvalidSpaceID, model.ErrFieldTextNotFilled,
			model.ErrFieldUserNotFilled, model.ErrFieldTypeNotFilled,
		}

		if errorsIn(err, errs) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		// внутренняя ошибка
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	req.ID = uuid.New()

	err = s.space.CreateNote(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusAccepted, map[string]string{"request_id": req.ID.String()})
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
//	    @Param        id   path      uuid  true  "ID пространства"
//		@Success		200 {object}    []model.Note
//		@Success		200 {object}    []model.GetNote
//		@Failure		400	{object}	map[string]string "Невалидный запрос"
//		@Failure		404                               "Пространства не существует"
//		@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//		@Router			/spaces/{id}/notes [get]
//
// ручка для получения всех заметок пользователя из его личного пространства
func (s *server) notesBySpaceID(c echo.Context) error {
	spaceIDStr := c.Param("id")

	spaceID, err := uuid.Parse(spaceIDStr)
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
		notes, err := s.space.GetAllNotesBySpaceIDFull(c.Request().Context(), spaceID)
		if err != nil {
			// у пользователя нет заметок - отдаем 404
			if errors.Is(err, api_errors.ErrNoNotesFoundBySpaceID) {
				return c.NoContent(http.StatusNotFound)
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
	notes, err := s.space.GetAllNotesBySpaceID(c.Request().Context(), spaceID)
	if err != nil {
		// у пользователя нет заметок - отдаем 404
		if errors.Is(err, api_errors.ErrNoNotesFoundBySpaceID) {
			return c.NoContent(http.StatusNotFound)
		}

		// пространство не существует - отдаем 404
		if errors.Is(err, api_errors.ErrSpaceNotExists) {
			return c.NoContent(http.StatusNotFound)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, notes)
}

//	@Summary		Запрос на обновление заметки
//	@Description	Запрос на обновление заметки с текстом. Создается в указанном пространстве
//	@Param			request	body	model.UpdateNoteRequest	true	"обновить заметку:\nуказать айди пользователя,\nайди его личного / совместного пространства,\nновый текст заметки,\nтип заметки: текст, фото, етс\nайди заметки, которую нужно обновить"
//	@Success		202 {object}    string             айди запроса для отслеживания
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/spaces/notes/update [patch]
//
// ручка для обновления заметки
func (s *server) updateNote(c echo.Context) error {
	var req model.UpdateNoteRequest

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	// валидируем данные
	if err := req.Validate(); err != nil {
		// ошибки запроса
		errs := []error{
			model.ErrInvalidSpaceID, model.ErrFieldTextNotFilled,
			model.ErrNoteIdNotFilled, model.ErrFieldUserNotFilled,
			model.ErrFieldTypeNotFilled, model.ErrUpdateNotTextNote,
		}

		if errorsIn(err, errs) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}
	}

	// проверяем, что в пространстве есть заметка с таким айди
	note, err := s.space.GetNoteByID(c.Request().Context(), req.NoteID)
	if err != nil {
		if errors.Is(err, api_errors.ErrNoteNotFound) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if note.SpaceID != req.SpaceID {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": api_errors.ErrNoteNotBelongsSpace.Error()})
	}

	// обновлять можно только текстовые заметки (пока)
	if note.Type != model.TextNoteType {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": model.ErrUpdateNotTextNote.Error()})
	}

	req.ID = uuid.New()

	if err := s.space.UpdateNote(c.Request().Context(), req); err != nil {
		// внутренняя ошибка
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// запрос принят в обработку
	return c.JSON(http.StatusAccepted, map[string]string{"request_id": req.ID.String()})
}

//	@Summary		Получить все типы заметок
//	@Description	Получить список всех типов заметок и их количество
//	@Param          id   path      string  true  "ID пространства"//
//	@Success		200 {object}    []model.NoteTypeResponse   массив с типами заметок и их количеством
//	@Failure		404	{object}	nil "Нет заметок"
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/spaces/{id}/notes/types [get]
//
// ручка для получения типов заметок
func (s *server) getNoteTypes(c echo.Context) error {
	spaceIDStr := c.Param("id")

	spaceID, err := uuid.Parse(spaceIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid space id parameter: %+v", err)})
	}

	types, err := s.space.GetNotesTypes(c.Request().Context(), spaceID)
	if err != nil {
		if errors.Is(err, api_errors.ErrNoNotesFoundBySpaceID) {
			return c.NoContent(http.StatusNotFound)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, types)
}

//	@Summary		Получить все заметки одного типа
//	@Description	Получить все заметки определенного типа: текстовые, фото, етс
//	@Param          id   path      string  true  "ID пространства"
//	@Param          type   path      string  true  "тип заметки: текст, фото, етс"
//	@Success		200 {object}    []model.GetNote   массив с типами заметок и их количеством
//	@Failure		404	{object}	nil "Нет заметок"
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/spaces/{id}/notes/{type} [get]
//
// ручка для заметок по типу
func (s *server) getNotesByType(c echo.Context) error {
	spaceIDStr := c.Param("id")
	noteType := c.Param("type")

	// валидируем запрос: тип должен быть одним из перечисленных
	switch noteType {
	case string(model.TextNoteType), string(model.PhotoNoteType):
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid note type: %s", noteType)})
	}

	spaceID, err := uuid.Parse(spaceIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid space id parameter: %+v", err)})
	}

	notes, err := s.space.GetNotesByType(c.Request().Context(), spaceID, model.NoteType(noteType))
	if err != nil {
		if errors.Is(err, api_errors.ErrNoNotesFoundByType) {
			return c.NoContent(http.StatusNotFound)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, notes)
}

//	@Summary		Получить все заметки по тексту
//	@Description	Получить все заметки с текстом среди указанного типа (по умолчанию: текстовые)
//	@Param          type   body      model.SearchNoteByTextRequest  true  "запрос на поиск по тексту"
//	@Success		200 {object}    []model.GetNote   массив с типами заметок и их количеством
//	@Failure		404	{object}	nil "Нет заметок"
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/spaces/notes/search/text [post]
//
// ручка для поиска заметок по тексту
func (s *server) searchNoteByText(c echo.Context) error {
	var req model.SearchNoteByTextRequest

	err := json.NewDecoder(c.Request().Body).Decode(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	if len(req.Type) > 0 {
		// валидируем запрос: тип должен быть одним из перечисленных
		switch req.Type {
		case string(model.TextNoteType), string(model.PhotoNoteType):
		default:
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid note type: %s", req.Type)})
		}
	}

	notes, err := s.space.SearchNoteByText(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, api_errors.ErrNoNotesFoundByText) {
			return c.NoContent(http.StatusNotFound)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, notes)
}
