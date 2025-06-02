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
//	@Router			/spaces/notes/create [post]
//
// ручка для создания заметки
func (h *handler) CreateNote(c echo.Context) error {
	var req rabbit.CreateNoteRequest

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
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
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		// внутренняя ошибка
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
//	    @Param        space_id   path      uuid  true  "ID пространства"
//		@Success		200 {object}    []model.Note
//		@Success		200 {object}    []model.GetNote
//		@Failure		400	{object}	map[string]string "Невалидный запрос"
//		@Failure		404                               "Пространства не существует"
//		@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//		@Router			/spaces/{space_id}/notes [get]
//
// ручка для получения всех заметок пользователя из его личного пространства
func (h *handler) NotesBySpaceID(c echo.Context) error {
	spaceID, err := getSpaceIDFromPath(c)
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

			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, notes)
}

//	@Summary		Запрос на обновление заметки
//	@Description	Запрос на обновление заметки с текстом. Создается в указанном пространстве
//	@Param			request	body	rabbit.UpdateNoteRequest	true	"обновить заметку:\nуказать айди пользователя,\nайди его личного / совместного пространства,\nновый текст заметки,\nтип заметки: текст, фото, етс\nайди заметки, которую нужно обновить"
//	@Success		202 {object}    string             айди запроса для отслеживания
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/spaces/notes/update [patch]
//
// ручка для обновления заметки
func (h *handler) UpdateNote(c echo.Context) error {
	var req rabbit.UpdateNoteRequest

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	req.ID = uuid.New()
	req.Created = time.Now().In(time.UTC).Unix()
	req.Operation = rabbit.UpdateOp

	// проверяем, что в пространстве есть заметка с таким айди
	note, err := h.space.GetNoteByID(c.Request().Context(), req.NoteID)
	if err != nil {
		if errors.Is(err, api_errors.ErrNoteNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}

		// ошибки запроса
		errs := []error{
			model.ErrInvalidSpaceID, model.ErrFieldTextNotFilled,
			model.ErrFieldUserNotFilled, model.ErrFieldTypeNotFilled,
			model.ErrUpdateNotTextNote,
		}

		if errorsIn(err, errs) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		// ошибку про поле created выше не проверяем, т.к. это внутренняя ошибка сервера, а не клиента
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if note.SpaceID != req.SpaceID {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": api_errors.ErrNoteNotBelongsSpace.Error()})
	}

	// обновлять можно только текстовые заметки (пока)
	if note.Type != model.TextNoteType {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": model.ErrUpdateNotTextNote.Error()})
	}

	if err := h.space.UpdateNote(c.Request().Context(), req); err != nil {
		// внутренняя ошибка
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// запрос принят в обработку
	return c.JSON(http.StatusAccepted, map[string]string{"request_id": req.ID.String()})
}

//	@Summary		Получить все типы заметок
//	@Description	Получить список всех типов заметок и их количество
//	@Param          space_id   path      string  true  "ID пространства"//
//	@Success		200 {object}    []model.NoteTypeResponse   массив с типами заметок и их количеством
//	@Failure		404	{object}	nil "Нет заметок"
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/spaces/{space_id}/notes/types [get]
//
// ручка для получения типов заметок
func (h *handler) GetNoteTypes(c echo.Context) error {
	spaceID, err := getSpaceIDFromPath(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid space id parameter: %+v", err)})
	}

	types, err := h.space.GetNotesTypes(c.Request().Context(), spaceID)
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
//	@Param          space_id   path      string  true  "ID пространства"
//	@Param          type   path      string  true  "тип заметки: текст, фото, етс"
//	@Success		200 {object}    []model.GetNote   массив с типами заметок и их количеством
//	@Failure		404	{object}	nil "Нет заметок"
//	@Failure		400	{object}	map[string]string "Невалидный запрос"
//	@Failure		500	{object}	map[string]string "Внутренняя ошибка"
//	@Router			/spaces/{space_id}/notes/{type} [get]
//
// ручка для заметок по типу
func (h *handler) GetNotesByType(c echo.Context) error {
	noteType := c.Param("type")

	// валидируем запрос: тип должен быть одним из перечисленных
	switch noteType {
	case string(model.TextNoteType), string(model.PhotoNoteType):
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid note type: %s", noteType)})
	}

	spaceID, err := getSpaceIDFromPath(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid space id parameter: %+v", err)})
	}

	notes, err := h.space.GetNotesByType(c.Request().Context(), spaceID, model.NoteType(noteType))
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
func (h *handler) SearchNoteByText(c echo.Context) error {
	var req model.SearchNoteByTextRequest

	err := json.NewDecoder(c.Request().Body).Decode(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	if len(req.Type) > 0 {
		// валидируем запрос: тип должен быть одним из перечисленных
		switch req.Type {
		case model.TextNoteType, model.PhotoNoteType:
		default:
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid note type: %s", req.Type)})
		}
	}

	notes, err := h.space.SearchNoteByText(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, api_errors.ErrNoNotesFoundByText) {
			return c.NoContent(http.StatusNotFound)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid space id parameter: %+v", err)})
	}

	noteID, err := getNoteIDFromPath(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid note id parameter: %+v", err)})
	}

	// проверяем, что пространство существует
	_, err = h.space.GetSpaceByID(c.Request().Context(), spaceID)
	if err != nil {
		if errors.Is(err, api_errors.ErrSpaceNotExists) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// проверяем, что в пространстве есть заметка с таким айди
	note, err := h.space.GetNoteByID(c.Request().Context(), noteID)
	if err != nil {
		// заметки не существует в принципе
		if errors.Is(err, api_errors.ErrNoteNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"bad request": err.Error()})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// заметка не из этого пространства
	if note.SpaceID != spaceID {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": api_errors.ErrNoteNotBelongsSpace.Error()})
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
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": api_errors.ErrNoteNotBelongsSpace.Error()})
		}

		// внутренняя ошибка / ошибка валидации
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusAccepted, map[string]string{"req_id": req.ID.String()})
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
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("invalid space id parameter: %+v", err)})
	}

	// проверяем, что пространство существует
	_, err = h.space.GetSpaceByID(c.Request().Context(), spaceID)
	if err != nil {
		if errors.Is(err, api_errors.ErrSpaceNotExists) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	req := rabbit.DeleteAllNotesRequest{
		ID:        uuid.New(),
		SpaceID:   spaceID,
		Created:   time.Now().In(time.UTC).Unix(),
		Operation: rabbit.DeleteAllOp,
	}

	if err := h.space.DeleteAllNotes(c.Request().Context(), req); err != nil {
		// внутренняя ошибка / ошибка валидации
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusAccepted, map[string]string{"req_id": req.ID.String()})
}

func getSpaceIDFromPath(c echo.Context) (uuid.UUID, error) {
	spaceIDStr := c.Param("space_id")

	return uuid.Parse(spaceIDStr)
}

func getNoteIDFromPath(c echo.Context) (uuid.UUID, error) {
	noteIDStr := c.Param("note_id")

	return uuid.Parse(noteIDStr)
}
