package v0

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *handler) CreateSpace(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"bad request": "token not found"})
	}

	token, err := h.auth.CheckToken(authHeader)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"bad request": err.Error()})
	}

	payload, ok := h.auth.GetPayload(token)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "payload in token not found"})
	}

	userIDStr, ok := payload["user_id"]
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"bad request": "user id not found in payload"})
	}

	userID := int64(userIDStr.(float64))

	err = h.user.CheckUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"bad request": err.Error()})
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
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusAccepted, map[string]string{"req_id": req.ID.String()})
}
