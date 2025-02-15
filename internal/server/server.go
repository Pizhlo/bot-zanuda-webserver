package server

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type server struct {
	addr string
	e    *echo.Echo
}

func New(addr string) *server {
	return &server{addr: addr}
}

func (s *server) Serve() error {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/health", nil)

	s.e = e

	return e.Start(s.addr)
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}
