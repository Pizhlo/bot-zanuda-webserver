package server

import (
	"context"
	"errors"
	"strings"

	_ "webserver/docs" // docs is generated by Swag CLI, you have to import it.

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Server struct {
	addr string
	e    *echo.Echo

	api struct {
		h0 handler
	}
}

type ServerOption func(*Server)

func WithAddr(addr string) ServerOption {
	return func(s *Server) {
		s.addr = addr
	}
}

func WithHandler(handler handler) ServerOption {
	return func(s *Server) {
		s.api.h0 = handler
	}
}

func New(opts ...ServerOption) (*Server, error) {
	server := &Server{}

	for _, opt := range opts {
		opt(server)
	}

	if server.addr == "" {
		return nil, errors.New("addr is required")
	}

	if server.api.h0 == nil {
		return nil, errors.New("handler is required")
	}

	return server, nil
}

//go:generate mockgen -source ./server.go -destination=./mocks/server.go -package=mocks
type handler interface {
	spaceHandler
	noteHandler
	middlewareHandler
	healthHandler
}

type spaceHandler interface {
	CreateSpace(c echo.Context) error
	AddParticipant(c echo.Context) error
}

type noteHandler interface {
	CreateNote(c echo.Context) error
	NotesBySpaceID(c echo.Context) error
	UpdateNote(c echo.Context) error
	GetNoteTypes(c echo.Context) error
	GetNotesByType(c echo.Context) error
	SearchNoteByText(c echo.Context) error
	DeleteNote(c echo.Context) error
	DeleteAllNotes(c echo.Context) error
}

type healthHandler interface {
	Health(c echo.Context) error
}

type middlewareHandler interface {
	ValidateNoteRequest(next echo.HandlerFunc) echo.HandlerFunc
	Auth(next echo.HandlerFunc) echo.HandlerFunc
	WrapNetHTTP(next echo.HandlerFunc) echo.HandlerFunc
}

func (s *Server) Start() error {
	if len(s.e.Routes()) > 0 {
		return s.e.Start(s.addr)
	}

	return errors.New("no routes initialized")
}

func (s *Server) CreateRoutes() error {
	e := echo.New()

	skipper := func(c echo.Context) bool {
		return strings.Contains(c.Request().URL.Path, "swagger")
	}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: skipper,
	}))

	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{Skipper: skipper}))

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Use(echoprometheus.NewMiddleware("webserver")) // adds middleware to gather metrics
	e.GET("/metrics", echoprometheus.NewHandler())   // adds route to serve gathered metrics

	api := e.Group("api/")

	// v0
	apiv0 := api.Group("v0/")

	apiv0.GET("health", s.api.h0.Health)

	spaces := apiv0.Group("spaces")

	// spaces
	spaces.POST("/create", s.api.h0.CreateSpace, s.api.h0.Auth, s.api.h0.WrapNetHTTP)                        // создать пространство
	spaces.POST("/:space_id/participants/add", s.api.h0.AddParticipant, s.api.h0.Auth, s.api.h0.WrapNetHTTP) // добавить участника в пространство

	// notes
	spaces.GET("/:space_id/notes", s.api.h0.NotesBySpaceID, s.api.h0.WrapNetHTTP)

	// создание, обновление, удаление
	spaces.POST("/notes/create", s.api.h0.CreateNote, s.api.h0.ValidateNoteRequest, s.api.h0.WrapNetHTTP)
	spaces.PATCH("/notes/update", s.api.h0.UpdateNote, s.api.h0.ValidateNoteRequest, s.api.h0.WrapNetHTTP)
	spaces.DELETE("/:space_id/notes/:note_id/delete", s.api.h0.DeleteNote, s.api.h0.WrapNetHTTP)
	spaces.DELETE("/:space_id/notes/delete_all", s.api.h0.DeleteAllNotes, s.api.h0.WrapNetHTTP) // удалить все заметки

	// типы заметок
	spaces.GET("/:space_id/notes/types", s.api.h0.GetNoteTypes, s.api.h0.WrapNetHTTP)   // получить, какие есть типы заметок
	spaces.GET("/:space_id/notes/:type", s.api.h0.GetNotesByType, s.api.h0.WrapNetHTTP) // получить все заметки одного типа

	// поиск
	spaces.POST("/notes/search/text", s.api.h0.SearchNoteByText, s.api.h0.WrapNetHTTP) // по тексту

	s.e = e

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}
