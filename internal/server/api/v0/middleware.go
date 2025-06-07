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
	"github.com/labstack/echo/v4"
)

// ValidateNoteRequest производит валидацию запросов на создание и обновление заметки.
// Проверяет: что пользователь существует, что пространство существует, что пользователь состоит в пространстве.
func (h *handler) ValidateNoteRequest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var note rabbit.UpdateNoteRequest

		// нам нужно сохранить тело запроса для последующей обработки в хендлерах
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(body, &note)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
		}

		// проверяем, что пользователь существует
		exists, err := h.user.CheckUser(c.Request().Context(), note.UserID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		if !exists {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": "user not found"})
		}

		// проверяем, что пространство существует
		if _, err := h.space.GetSpaceByID(c.Request().Context(), note.SpaceID); err != nil {
			if errors.Is(err, api_errors.ErrSpaceNotBelongsUser) || errors.Is(err, api_errors.ErrSpaceNotExists) {
				return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
			}

			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// проверяем, что пользователь состоит в пространстве (сюда потом еще добавится проверка на права)
		exists, err = h.space.IsUserInSpace(c.Request().Context(), note.UserID, note.SpaceID)
		if err != nil {
			if errors.Is(err, api_errors.ErrSpaceNotBelongsUser) || errors.Is(err, api_errors.ErrSpaceNotExists) {
				return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
			}

			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		if !exists {
			return c.JSON(http.StatusBadRequest, map[string]string{"bad request": "user not in space"})
		}

		// Восстанавливаем тело запроса, чтобы его можно было прочитать в хендлере
		c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

		return next(c)
	}
}

func (h *handler) Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
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
			return c.JSON(http.StatusUnauthorized, map[string]string{"bad request": "payload in token not found"})
		}

		userID, err := getUserIDFromToken(payload)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"bad request": err.Error()})
		}

		expired, err := getExpiredFromToken(payload)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"bad request": err.Error()})
		}

		if expired < time.Now().Unix() {
			return c.JSON(http.StatusUnauthorized, map[string]string{"bad request": "token expired"})
		}

		exists, err := h.user.CheckUser(c.Request().Context(), userID)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"bad request": err.Error()})
		}

		if !exists {
			return c.JSON(http.StatusUnauthorized, map[string]string{"bad request": fmt.Sprintf("user %d not found", userID)})
		}

		c.Request().Header.Set("user_id", strconv.FormatInt(userID, 10)) // чтобы в хендлерах был доступ к айди пользователя

		return next(c)
	}
}

func getUserIDFromToken(payload jwt.MapClaims) (int64, error) {
	userIDStr, ok := payload["user_id"]
	if !ok {
		return 0, errors.New("user id not found in payload")
	}

	return int64(userIDStr.(float64)), nil
}

func getExpiredFromToken(payload jwt.MapClaims) (int64, error) {
	expiredAny, ok := payload["expired"]
	if !ok {
		return 0, errors.New("expired not found in payload")
	}

	return int64(expiredAny.(float64)), nil
}
