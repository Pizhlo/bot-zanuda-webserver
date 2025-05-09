package server

import (
	"fmt"
	"net/http"
	"testing"
	"webserver/mocks"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := mocks.NewMockhandler(ctrl)

	cfg, err := NewConfig("addr", h)
	require.NoError(t, err)

	server := New(cfg)
	assert.NotNil(t, server)
	assert.Equal(t, cfg.Address, server.addr)
	assert.Equal(t, cfg.HandlerV0, server.api.h0)
}

func TestCreateRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := mocks.NewMockhandler(ctrl)

	cfg, err := NewConfig(":8080", h)
	require.NoError(t, err)

	server := New(cfg)

	err = server.CreateRoutes()
	require.NoError(t, err)

	routes := server.e.Routes()

	expectedRoutes := []*echo.Route{
		{
			Method: http.MethodGet,
			Path:   "/api/v0/health",
			Name:   "webserver/internal/server.handler.Health-fm",
		},
		{
			Method: http.MethodGet,
			Path:   "/swagger/*",
			Name:   "github.com/swaggo/echo-swagger.EchoWrapHandler.func1",
		},
		{
			Method: http.MethodGet,
			Path:   "/api/v0/spaces/:id/notes",
			Name:   "webserver/internal/server.handler.NotesBySpaceID-fm",
		},
		{
			Method: http.MethodPost,
			Path:   "/api/v0/spaces/notes/create",
			Name:   "webserver/internal/server.handler.CreateNote-fm",
		},
		{
			Method: http.MethodPatch,
			Path:   "/api/v0/spaces/notes/update",
			Name:   "webserver/internal/server.handler.UpdateNote-fm",
		},
		{
			Method: http.MethodDelete,
			Path:   "/api/v0/spaces/:space_id/notes/:note_id/delete",
			Name:   "webserver/internal/server.handler.DeleteNote-fm",
		},
		{
			Method: http.MethodDelete,
			Path:   "/api/v0/spaces/:space_id/notes/delete_all",
			Name:   "webserver/internal/server.handler.DeleteAllNotes-fm",
		},
		{
			Method: http.MethodGet,
			Path:   "/api/v0/spaces/:id/notes/types",
			Name:   "webserver/internal/server.handler.GetNoteTypes-fm",
		},
		{
			Method: http.MethodGet,
			Path:   "/api/v0/spaces/:id/notes/:type",
			Name:   "webserver/internal/server.handler.GetNotesByType-fm",
		},
		{
			Method: http.MethodPost,
			Path:   "/api/v0/spaces/notes/search/text",
			Name:   "webserver/internal/server.handler.SearchNoteByText-fm",
		},
	}

	assert.Equal(t, len(expectedRoutes), len(routes))

	actualRoutesMap := routesMap(routes)

	// not found routes
	notFound := []string{}

	// path: method
	expRoutesMap := routesMap(expectedRoutes)

	for expectedPath, expectedMethod := range expRoutesMap {
		if actualMethod, found := actualRoutesMap[expectedPath]; !found {
			notFound = append(notFound, expectedPath)
		} else {
			assert.Equal(t, expectedMethod, actualMethod, fmt.Sprintf("methods not equal for path '%s'", expectedPath))
		}
	}

	if len(notFound) > 0 {
		t.Errorf("not found paths: %+v", notFound)
	}
}

func routesMap(routes []*echo.Route) map[string]string {
	res := map[string]string{}

	for _, r := range routes {
		res[r.Path] = r.Method
	}

	return res
}
