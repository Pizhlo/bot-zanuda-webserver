package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body io.Reader) *http.Response {

	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	req.Close = true
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("User-Agent", "PostmanRuntime/7.32.3")
	require.NoError(t, err)

	ts.Client()

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	return resp
}

func runTestServer(server *server) (*echo.Echo, error) {
	e := echo.New()

	server.e = e

	e.GET("/health", server.health)
	spaces := e.Group("spaces")

	// notes
	spaces.GET("/:id/notes", server.notesBySpaceID)
	spaces.POST("/notes/create", server.createNote, server.validateNoteRequest)
	spaces.PATCH("/notes/update", server.updateNote, server.validateNoteRequest)
	spaces.GET("/:id/notes/types", server.getNoteTypes)
	spaces.GET("/:id/notes/:type", server.getNotesByType)
	spaces.POST("/notes/search/text", server.searchNoteByText)

	return e, nil
}
