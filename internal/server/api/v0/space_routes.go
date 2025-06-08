package v0

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// @Summary		Запрос на создание пространства
// @Description	Запрос на создание пространства
// @Param			request	body	rabbit.CreateSpaceRequest	true	"создать пространство:\nуказать айди пользователя,\nайди его личного / совместного пространства,\nтекст заметки\nтип заметки: текстовый, фото, видео, и т.п."
// @Success		202 {object}    string             айди запроса для отслеживания
// @Failure		400	{object}	map[string]string "Невалидный запрос"
// @Failure		401	{object}	map[string]string "Невалидный токен"
// @Failure		500	{object}	map[string]string "Внутренняя ошибка"
// @Router			/api/v0/spaces/create [post]
func (h *Handler) CreateSpace(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
	}

	var req rabbit.CreateSpaceRequest

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
	req.UserID = userID

	if err := h.space.CreateSpace(c.Request().Context(), req); err != nil {
		if errors.Is(err, model.ErrFieldNameNotFilled) {
			return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
		}

		// ошибку про поле created выше не проверяем, т.к. это внутренняя ошибка сервера, а не клиента
		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	return sendRequestID(c, req.ID)
}

// @Summary		Запрос на добавление участника в пространство
// @Description	Запрос на добавление участника в пространство
// @Param          space_id   path      string  true  "ID пространства"
// @Param		request	body	rabbit.AddParticipantRequest	true	"добавить участника в пространство:\nуказать айди пользователя,\nайди совместного пространства"
// @Success		202 {object}    string             айди запроса для отслеживания
// @Failure		400	{object}	map[string]string "Невалидный запрос"
// @Failure		401	{object}	map[string]string "Невалидный токен"
// @Failure		500	{object}	map[string]string "Внутренняя ошибка"
// @Router			/api/v0/spaces/{space_id}/participants/add [post]
func (h *Handler) AddParticipant(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
	}

	spaceID, err := getSpaceIDFromPath(c)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
	}

	var req rabbit.AddParticipantRequest

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
	}

	// нельзя добавить самого себя
	if req.Participant == userID {
		return api_errors.NewHTTPError(http.StatusBadRequest, "you can't add yourself as a participant", nil)
	}

	// 400 пользователя (которого пригласили) не существует
	// проверяем, что существует пользователь, которого добавляем
	exists, err := h.user.CheckUser(c.Request().Context(), req.Participant)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusBadRequest, err.Error(), err)
	}

	if !exists {
		return api_errors.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("user %d not found", req.Participant), nil)
	}

	// 400 не найдено совместное пространство с таким айди
	// проверка, что пространство не личное (и что оно существует)
	isPersonal, err := h.space.IsSpacePersonal(c.Request().Context(), spaceID)
	if err != nil {
		if errors.Is(err, api_errors.ErrSpaceNotExists) {
			return api_errors.NewHTTPError(http.StatusBadRequest, "space not found", err)
		}

		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	// 400 пространство личное
	if isPersonal {
		return api_errors.NewHTTPError(http.StatusBadRequest, "personal space", nil)
	}

	// 400 пользователь (который приглашает) не состоит в пространстве
	exists, err = h.space.IsUserInSpace(c.Request().Context(), userID, spaceID)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	if !exists {
		return api_errors.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("user %d not in space", userID), nil)
	}

	// 400 приглашенный пользователь уже в пространстве
	exists, err = h.space.IsUserInSpace(c.Request().Context(), req.Participant, spaceID)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	if exists {
		return api_errors.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("user %d already in space", req.Participant), nil)
	}

	// проверяем, что пользователь еще не приглашен в пространство
	// 400 уже существует такое приглашение (пользователь А уже пригласил пользователя В)
	exists, err = h.space.CheckInvitation(c.Request().Context(), userID, req.Participant, spaceID)
	if err != nil {
		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	if exists {
		return api_errors.NewHTTPError(http.StatusBadRequest, "invitation already exists", nil)
	}

	req.ID = uuid.New()
	req.Created = time.Now().In(time.UTC).Unix()
	req.Operation = rabbit.AddParticipantOp
	req.UserID = userID
	req.SpaceID = spaceID

	if err := h.space.AddParticipant(c.Request().Context(), req); err != nil {
		return api_errors.NewHTTPError(http.StatusInternalServerError, err.Error(), err)
	}

	return sendRequestID(c, req.ID)
}

func getUserID(c echo.Context) (int64, error) {
	userIDStr := c.Request().Header.Get("user_id")
	if userIDStr == "" {
		return 0, errors.New("user id not found in request header")
	}

	return strconv.ParseInt(userIDStr, 10, 64)
}
