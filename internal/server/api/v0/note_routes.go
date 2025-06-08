package v0

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	api_errors "webserver/internal/errors"
)

//	@Summary		Запрос на создание заметки
//	@Description	Запрос на создание заметки с текстом. Создается в указанном пространстве
//	@Param			request	body	rabbit.CreateNoteRequest	true	"создать заметку:\nуказать айди пользователя,\nайди его личного / совместного пространства,\nтекст заметки\nтип заметки: текстовый, фото, видео, и т.п."
//	@Success		202 {object}    string             айди запроса для отслеживания
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/api/v0/spaces/notes/create [post]
//
// ручка для создания заметки
func (h *handler) CreateNote(c echo.Context) error {
	var req rabbit.CreateNoteRequest

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
	}

	req.ID = uuid.New()
	req.Created = time.Now().In(time.UTC).Unix()
	req.Operation = rabbit.CreateOp

	err = h.space.CreateNote(c.Request().Context(), req)
	if err != nil {
		// ошибки запроса
		errs := []error{
			model.ErrInvalidSpaceID, model.ErrFieldTextNotFilled,
			model.ErrFieldUserNotFilled, model.ErrFieldTypeNotFilled,
		}

		if errorsIn(err, errs) {
			return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
		}

		// внутренняя ошибка
		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	return sendRequestID(c, req.ID)
}

func errorsIn(target error, errs []error) bool {
	for _, err := range errs {
		if errors.Is(target, err) {
			return true
		}
	}
	return false
}

//			@Summary		Запрос на получение всех заметок
//			@Description	Запрос на получение всех заметок из личного пространства пользователя
//	     @Param          space_id   path      string  true  "ID пространства"
//			@Success		200 {object}    []model.Note
//			@Success		200 {object}    []model.GetNote
//			@Failure		400	{object}	map[string]string "Невалидный запрос"
//			@Failure		404                               "Пространства не существует"
//			@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//			@Router			/api/v0/spaces/{space_id}/notes [get]
//
// ручка для получения всех заметок пользователя из его личного пространства
func (h *handler) NotesBySpaceID(c echo.Context) error {
	spaceID, err := getSpaceIDFromPath(c)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid space id parameter: %+v", err), err)
	}

	// предоставлять ли полную инф-ю о пользователе, который создал заметку
	fullUserParam := c.QueryParam("full_user")

	var fullUser bool // по умолчанию false

	if len(fullUserParam) > 0 {
		var err error
		fullUser, err = strconv.ParseBool(fullUserParam)
		if err != nil {
			return api_errors.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid full user parameter: %+v", err), err)
		}
	}

	// получение заметок в полном режиме
	if fullUser {
		notes, err := h.space.GetAllNotesBySpaceIDFull(c.Request().Context(), spaceID)
		if err != nil {
			// у пользователя нет заметок - отдаем 404
			if errors.Is(err, api_errors.ErrNoNotesFoundBySpaceID) {
				return c.NoContent(http.StatusNotFound)
			}

			// пространство не существует - отдаем 404
			if errors.Is(err, api_errors.ErrSpaceNotExists) {
				return c.NoContent(http.StatusNotFound)
			}

			return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
		}

		return c.JSON(http.StatusOK, notes)
	}

	// получение заметок в кратком режиме
	notes, err := h.space.GetAllNotesBySpaceID(c.Request().Context(), spaceID)
	if err != nil {
		// у пользователя нет заметок - отдаем 404
		if errors.Is(err, api_errors.ErrNoNotesFoundBySpaceID) {
			return c.NoContent(http.StatusNotFound)
		}

		// пространство не существует - отдаем 404
		if errors.Is(err, api_errors.ErrSpaceNotExists) {
			return c.NoContent(http.StatusNotFound)
		}

		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	return c.JSON(http.StatusOK, notes)
}

//	@Summary		Запрос на обновление заметки
//	@Description	Запрос на обновление заметки с текстом. Создается в указанном пространстве
//	@Param			request	body	rabbit.UpdateNoteRequest	true	"обновить заметку:\nуказать айди пользователя,\nайди его личного / совместного пространства,\nновый текст заметки,\nтип заметки: текст, фото, етс\nайди заметки, которую нужно обновить"
//	@Success		202 {object}    string             айди запроса для отслеживания
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/api/v0/spaces/notes/update [patch]
//
// ручка для обновления заметки
func (h *handler) UpdateNote(c echo.Context) error {
	var req rabbit.UpdateNoteRequest

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
	}

	req.ID = uuid.New()
	req.Created = time.Now().In(time.UTC).Unix()
	req.Operation = rabbit.UpdateOp

	// проверяем, что в пространстве есть заметка с таким айди
	note, err := h.space.GetNoteByID(c.Request().Context(), req.NoteID)
	if err != nil {
		if errors.Is(err, api_errors.ErrNoteNotFound) {
			return api_errors.NewHTTPError(http.StatusNotFound, err.Error(), err)
		}

		// ошибки запроса
		errs := []error{
			model.ErrInvalidSpaceID, model.ErrFieldTextNotFilled,
			model.ErrFieldUserNotFilled, model.ErrFieldTypeNotFilled,
			model.ErrUpdateNotTextNote,
		}

		if errorsIn(err, errs) {
			return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
		}

		// ошибку про поле created выше не проверяем, т.к. это внутренняя ошибка сервера, а не клиента
		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	if note.SpaceID != req.SpaceID {
		return api_errors.NewHTTPError(http.StatusBadRequest, api_errors.ErrNoteNotBelongsSpace.Error(), nil)
	}

	// обновлять можно только текстовые заметки (пока)
	if note.Type != model.TextNoteType {
		return api_errors.NewHTTPError(http.StatusBadRequest, model.ErrUpdateNotTextNote.Error(), nil)
	}

	if err := h.space.UpdateNote(c.Request().Context(), req); err != nil {
		// внутренняя ошибка
		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	// запрос принят в обработку
	return sendRequestID(c, req.ID)
}

//	@Summary		Получить все типы заметок
//	@Description	Получить список всех типов заметок и их количество
//	@Param          space_id   path      string  true  "ID пространства"//
//	@Success		200 {object}    []model.NoteTypeResponse   массив с типами заметок и их количеством
//	@Failure		404	{object}	nil "Нет заметок"
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/api/v0/spaces/{space_id}/notes/types [get]
//
// ручка для получения типов заметок
func (h *handler) GetNoteTypes(c echo.Context) error {
	spaceID, err := getSpaceIDFromPath(c)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid space id parameter: %+v", err), err)
	}

	types, err := h.space.GetNotesTypes(c.Request().Context(), spaceID)
	if err != nil {
		if errors.Is(err, api_errors.ErrNoNotesFoundBySpaceID) {
			return c.NoContent(http.StatusNotFound)
		}

		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	return c.JSON(http.StatusOK, types)
}

//	@Summary		Получить все заметки одного типа
//	@Description	Получить все заметки определенного типа: текстовые, фото, етс
//	@Param          space_id   path      string  true  "ID пространства"
//	@Param          type   path      string  true  "тип заметки: текст, фото, етс"
//	@Success		200 {object}    []model.GetNote   массив с типами заметок и их количеством
//	@Failure		404	{object}	nil "Нет заметок"
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/api/v0/spaces/{space_id}/notes/{type} [get]
//
// ручка для заметок по типу
func (h *handler) GetNotesByType(c echo.Context) error {
	noteType := c.Param("type")

	// валидируем запрос: тип должен быть одним из перечисленных
	switch noteType {
	case string(model.TextNoteType), string(model.PhotoNoteType):
	default:
		return api_errors.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid note type: %s", noteType), nil)
	}

	spaceID, err := getSpaceIDFromPath(c)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid space id parameter: %+v", err), err)
	}

	notes, err := h.space.GetNotesByType(c.Request().Context(), spaceID, model.NoteType(noteType))
	if err != nil {
		if errors.Is(err, api_errors.ErrNoNotesFoundByType) {
			return c.NoContent(http.StatusNotFound)
		}

		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
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
func (h *handler) SearchNoteByText(c echo.Context) error {
	var req model.SearchNoteByTextRequest

	err := json.NewDecoder(c.Request().Body).Decode(&req)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
	}

	if len(req.Type) > 0 {
		// валидируем запрос: тип должен быть одним из перечисленных
		switch req.Type {
		case model.TextNoteType, model.PhotoNoteType:
		default:
			return api_errors.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid note type: %s", req.Type), nil)
		}
	}

	notes, err := h.space.SearchNoteByText(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, api_errors.ErrNoNotesFoundByText) {
			return c.NoContent(http.StatusNotFound)
		}

		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	return c.JSON(http.StatusOK, notes)
}

//	@Summary		Удалить заметку по айди
//	@Param          space_id   path      string  true  "айди пространства"
//	@Param          note_id   path      string  true  "айди заметки"
//	@Success		202 {object}    map[string]string "Айди запроса"
//	@Failure		400	{object}	map[string]string "Пространства не существует / в пространстве нет такой заметки"
//	@Failure		404	{object}	map[string]string "Заметка не найдена"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/spaces/{space_id}/notes/{note_id}/delete [delete]
//
// ручка для удаления заметки по id
func (h *handler) DeleteNote(c echo.Context) error {
	spaceID, err := getSpaceIDFromPath(c)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid space id parameter: %+v", err), err)
	}

	noteID, err := getNoteIDFromPath(c)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid note id parameter: %+v", err), err)
	}

	// проверяем, что пространство существует
	_, err = h.space.GetSpaceByID(c.Request().Context(), spaceID)
	if err != nil {
		if errors.Is(err, api_errors.ErrSpaceNotExists) {
			return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
		}

		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	// проверяем, что в пространстве есть заметка с таким айди
	note, err := h.space.GetNoteByID(c.Request().Context(), noteID)
	if err != nil {
		// заметки не существует в принципе
		if errors.Is(err, api_errors.ErrNoteNotFound) {
			return api_errors.NewHTTPError(http.StatusNotFound, err.Error(), err)
		}

		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	// заметка не из этого пространства
	if note.SpaceID != spaceID {
		return api_errors.NewHTTPError(http.StatusBadRequest, api_errors.ErrNoteNotBelongsSpace.Error(), nil)
	}

	req := rabbit.DeleteNoteRequest{
		ID:        uuid.New(),
		SpaceID:   spaceID,
		NoteID:    noteID,
		Created:   time.Now().In(time.UTC).Unix(),
		Operation: rabbit.DeleteOp,
	}

	err = h.space.DeleteNote(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, model.ErrInvalidSpaceID) {
			return api_errors.NewHTTPError(http.StatusBadRequest, api_errors.ErrNoteNotBelongsSpace.Error(), nil)
		}

		// внутренняя ошибка / ошибка валидации
		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	return sendRequestID(c, req.ID)
}

// @Summary		Удалить все заметки в пространстве
// @Description	Удалить все заметки в пространстве
// @Param          space_id   path      string  true  "айди пространства"
// @Success		202 {object}    map[string]string "Айди запроса"
// @Failure		400	{object}	map[string]string "Пространства не существует"
// @Failure		500	{object}	map[string]string "Внутренняя ошибка"
// @Router			/spaces/{space_id}/notes/delete [delete]
func (h *handler) DeleteAllNotes(c echo.Context) error {
	spaceID, err := getSpaceIDFromPath(c)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid space id parameter: %+v", err), err)
	}

	// проверяем, что пространство существует
	_, err = h.space.GetSpaceByID(c.Request().Context(), spaceID)
	if err != nil {
		if errors.Is(err, api_errors.ErrSpaceNotExists) {
			return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
		}

		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	req := rabbit.DeleteAllNotesRequest{
		ID:        uuid.New(),
		SpaceID:   spaceID,
		Created:   time.Now().In(time.UTC).Unix(),
		Operation: rabbit.DeleteAllOp,
	}

	if err := h.space.DeleteAllNotes(c.Request().Context(), req); err != nil {
		// внутренняя ошибка / ошибка валидации
		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	return sendRequestID(c, req.ID)
}

func getSpaceIDFromPath(c echo.Context) (uuid.UUID, error) {
	spaceIDStr := c.Param("space_id")

	return uuid.Parse(spaceIDStr)
}

func getNoteIDFromPath(c echo.Context) (uuid.UUID, error) {
	noteIDStr := c.Param("note_id")

	return uuid.Parse(noteIDStr)
}

func sendRequestID(c echo.Context, reqID uuid.UUID) error {
	return c.JSON(http.StatusAccepted, map[string]string{"request_id": reqID.String()})
}
