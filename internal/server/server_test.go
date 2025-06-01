package server

import (
	"errors"
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

	tests := []struct {
		name string
		opts []ServerOption
		want *server
		err  error
	}{
		{
			name: "positive case",
			opts: []ServerOption{
				WithAddr(":8080"),
				WithHandler(mocks.NewMockhandler(ctrl)),
			},
			want: &server{
				addr: ":8080",
				api: struct {
					h0 handler
				}{h0: mocks.NewMockhandler(ctrl)},
			},
			err: nil,
		},
		{
			name: "error case: addr is required",
			opts: []ServerOption{
				WithHandler(mocks.NewMockhandler(ctrl)),
			},
			err: errors.New("addr is required"),
		},
		{
			name: "error case: handler is required",
			opts: []ServerOption{
				WithAddr(":8080"),
			},
			err: errors.New("handler is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := New(tt.opts...)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
				assert.Nil(t, server)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, server)
			}
		})
	}
}

func TestCreateRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := mocks.NewMockhandler(ctrl)

	server, err := New(
		WithAddr(":8080"),
		WithHandler(h),
	)
	require.NoError(t, err)

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
			Path:   "/api/v0/spaces/:space_id/notes",
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
			Path:   "/api/v0/spaces/:space_id/notes/types",
			Name:   "webserver/internal/server.handler.GetNoteTypes-fm",
		},
		{
			Method: http.MethodGet,
			Path:   "/api/v0/spaces/:space_id/notes/:type",
			Name:   "webserver/internal/server.handler.GetNotesByType-fm",
		},
		{
			Method: http.MethodPost,
			Path:   "/api/v0/spaces/notes/search/text",
			Name:   "webserver/internal/server.handler.SearchNoteByText-fm",
		},
		{
			Method: http.MethodPost,
			Path:   "/api/v0/spaces/create",
			Name:   "webserver/internal/server.handler.CreateSpace-fm",
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
