package v0

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"

	"github.com/labstack/echo/v4"
)

// ValidateNoteRequest производит валидацию запросов на создание и обновление заметки.
// Проверяет: что пользователь существует, что пространство существует, что пользователь состоит в пространстве.
func (h *handler) ValidateNoteRequest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var note model.UpdateNoteRequest

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
		if err := h.user.CheckUser(c.Request().Context(), note.UserID); err != nil {
			if errors.Is(err, api_errors.ErrUnknownUser) {
				return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
			}

			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// проверяем, что пространство существует
		if _, err := h.space.GetSpaceByID(c.Request().Context(), note.SpaceID); err != nil {
			if errors.Is(err, api_errors.ErrSpaceNotBelongsUser) || errors.Is(err, api_errors.ErrSpaceNotExists) {
				return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
			}

			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// проверяем, что пользователь состоит в пространстве (сюда потом еще добавится проверка на права)
		if err := h.space.IsUserInSpace(c.Request().Context(), note.UserID, note.SpaceID); err != nil {
			if errors.Is(err, api_errors.ErrSpaceNotBelongsUser) || errors.Is(err, api_errors.ErrSpaceNotExists) {
				return c.JSON(http.StatusBadRequest, map[string]string{"bad request": err.Error()})
			}

			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Восстанавливаем тело запроса, чтобы его можно было прочитать в хендлере
		c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

		return next(c)
	}
}
