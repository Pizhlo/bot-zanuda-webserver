package server

import (
	"time"

	"github.com/labstack/echo/v4"
)

// logging - middleware для логирования запросов.
// Логирует метод, URI, статус, хост, IP-адрес, длительность запроса, размер входящих и исходящих данных,
// пользовательский агент и протокол.
// Если в запросе произошла ошибка, то она также логируется.
func (s *Server) logging(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		s.logger.WithField("uri", c.Request().RequestURI).Info("HTTP request received")

		fields := map[string]interface{}{
			"method":     c.Request().Method,
			"uri":        c.Request().RequestURI,
			"status":     c.Response().Status,
			"host":       c.Request().Host,
			"remote_ip":  c.RealIP(),
			"bytes_in":   c.Request().ContentLength,
			"bytes_out":  c.Response().Size,
			"user_agent": c.Request().UserAgent(),
			"protocol":   c.Request().Proto,
		}

		err := next(c)

		fields["latency_ms"] = time.Since(start).Milliseconds()

		if err != nil {
			fields["error"] = err
			s.logger.WithFields(fields).Error("HTTP request processed with error")
			return err
		}

		s.logger.WithFields(fields).Info("HTTP request processed")

		return nil
	}
}
