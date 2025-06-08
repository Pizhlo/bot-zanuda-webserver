package v0

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	api_errors "webserver/internal/errors"
	"webserver/internal/model/rabbit"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// ValidateNoteRequest производит валидацию запросов на создание и обновление заметки.
// Проверяет: что пользователь существует, что пространство существует, что пользователь состоит в пространстве.
func (h *Handler) ValidateNoteRequest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// нам нужно сохранить тело запроса для последующей обработки в хендлерах
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		// Определяем тип запроса по URL
		var userID int64
		var spaceID uuid.UUID

		switch c.Path() {
		case "/api/v0/spaces/notes/create":
			var note rabbit.CreateNoteRequest
			if err := json.Unmarshal(body, &note); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			}
			userID = note.UserID
			spaceID = note.SpaceID
		case "/api/v0/spaces/notes/update":
			var note rabbit.UpdateNoteRequest
			if err := json.Unmarshal(body, &note); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			}
			userID = note.UserID
			spaceID = note.SpaceID
		default:
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "unsupported operation"})
		}

		// проверяем, что пользователь существует
		exists, err := h.user.CheckUser(c.Request().Context(), userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		if !exists {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": api_errors.ErrUnknownUser.Error()})
		}

		// проверяем, что пространство существует
		exists, err = h.space.IsSpaceExists(c.Request().Context(), spaceID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		if !exists {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": api_errors.ErrSpaceNotExists.Error()})
		}

		// проверяем, что пользователь состоит в пространстве (сюда потом еще добавится проверка на права)
		exists, err = h.space.IsUserInSpace(c.Request().Context(), userID, spaceID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		if !exists {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": api_errors.ErrSpaceNotBelongsUser.Error()})
		}

		// Восстанавливаем тело запроса
		c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

		return next(c)
	}
}

func (h *Handler) Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "token not found"})
		}

		token, err := h.auth.CheckToken(authHeader)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}

		payload, ok := h.auth.GetPayload(token)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": api_errors.ErrNoPayloadInToken.Error()})
		}

		userID, err := getUserIDFromToken(payload)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}

		expired, err := getExpiredFromToken(payload)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}

		if expired < time.Now().Unix() {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": api_errors.ErrTokenExpired.Error()})
		}

		exists, err := h.user.CheckUser(c.Request().Context(), userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		if !exists {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": fmt.Sprintf("user %d not found", userID)})
		}

		c.Request().Header.Set("user_id", strconv.FormatInt(userID, 10)) // чтобы в хендлерах был доступ к айди пользователя

		return next(c)
	}
}

// WrapNetHTTP вызывает следующий хендлер и обрабатывает его ответ.
func (h *Handler) WrapNetHTTP(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := next(c); err != nil {
			var httpErr *api_errors.HTTPError
			if errors.As(err, &httpErr) {
				// Если это наша кастомная HTTPError
				if httpErr.InnerError != nil {
					logrus.Errorf("Client Message: %s, Internal Error: %s. Status Code: %d", httpErr.Message, httpErr.InnerError, httpErr.Code)
				} else {
					logrus.Errorf("HTTP error: %d %s", httpErr.Code, httpErr.Message)
				}

				return c.JSON(httpErr.Code, map[string]string{"error": httpErr.Message})
			} else {
				// Если это другая, непредвиденная ошибка
				logrus.Errorf("Internal server error: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
			}
		}

		return nil
	}
}

func getUserIDFromToken(payload jwt.MapClaims) (int64, error) {
	userIDStr, ok := payload["user_id"]
	if !ok {
		return 0, api_errors.ErrUserNotFoundInPayload
	}

	return int64(userIDStr.(float64)), nil
}

func getExpiredFromToken(payload jwt.MapClaims) (int64, error) {
	expiredAny, ok := payload["expired"]
	if !ok {
		return 0, api_errors.ErrExpiredNotFoundInPayload
	}

	return int64(expiredAny.(float64)), nil
}
