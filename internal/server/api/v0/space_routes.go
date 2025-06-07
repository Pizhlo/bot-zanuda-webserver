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
// @Router			/spaces/create [post]
func (h *handler) CreateSpace(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	var req rabbit.CreateSpaceRequest

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
	req.UserID = userID

	if err := h.space.CreateSpace(c.Request().Context(), req); err != nil {
		if errors.Is(err, model.ErrFieldNameNotFilled) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		// ошибку про поле created выше не проверяем, т.к. это внутренняя ошибка сервера, а не клиента
		return sendInternalError(c, err)
	}

	return sendRequestID(c, req.ID)
}

func (h *handler) AddParticipant(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	spaceID, err := getSpaceIDFromPath(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	var req rabbit.AddParticipantRequest

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	// нельзя добавить самого себя
	if req.Participant == userID {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": "you can't add yourself as a participant"})
	}

	// 400 пользователя (которого пригласили) не существует
	// проверяем, что существует пользователь, которого добавляем
	exists, err := h.user.CheckUser(c.Request().Context(), req.Participant)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
	}

	if !exists {
		return c.JSON(http.StatusUnauthorized, map[string]string{"bad request": fmt.Sprintf("user %d not found", req.Participant)})
	}

	// 400 не найдено совместное пространство с таким айди
	// проверка, что пространство не личное (и что оно существует)
	isPersonal, err := h.space.IsSpacePersonal(c.Request().Context(), spaceID)
	if err != nil {
		if errors.Is(err, api_errors.ErrSpaceNotExists) {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": "space not found"})
		}

		return sendInternalError(c, err)
	}

	// 400 пространство личное
	if isPersonal {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": "personal space"})
	}

	// 400 пользователь (который приглашает) не состоит в пространстве
	exists, err = h.space.IsUserInSpace(c.Request().Context(), userID, spaceID)
	if err != nil {
		return sendInternalError(c, err)
	}

	if !exists {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("user %d not in space", userID)})
	}

	// 400 приглашенный пользователь уже в пространстве
	exists, err = h.space.IsUserInSpace(c.Request().Context(), req.Participant, spaceID)
	if err != nil {
		return sendInternalError(c, err)
	}

	if exists {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": fmt.Sprintf("user %d already in space", req.Participant)})
	}

	// проверяем, что пользователь еще не приглашен в пространство
	// 400 уже существует такое приглашение (пользователь А уже пригласил пользователя В)
	exists, err = h.space.CheckInvitation(c.Request().Context(), userID, req.Participant, spaceID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if exists {
		return c.JSON(http.StatusBadRequest, map[string]string{"bad request": "invitation already exists"})
	}

	req.ID = uuid.New()
	req.Created = time.Now().In(time.UTC).Unix()
	req.Operation = rabbit.AddParticipantOp
	req.UserID = userID
	req.SpaceID = spaceID

	if err := h.space.AddParticipant(c.Request().Context(), req); err != nil {
		return sendInternalError(c, err)
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
