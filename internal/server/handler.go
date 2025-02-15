package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// health необходим для проверки работоспособности сервера.
// Всегда отвечает 200 ОК
func (s *server) health(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}
