package v0

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

func runTestServer(h *Handler) (*echo.Echo, error) {
	e := echo.New()

	e.GET("/health", h.Health)
	spaces := e.Group("spaces")

	// notes
	spaces.GET("/:id/notes", h.NotesBySpaceID)
	spaces.POST("/notes/create", h.CreateNote, h.ValidateNoteRequest)
	spaces.PATCH("/notes/update", h.UpdateNote, h.ValidateNoteRequest)
	spaces.GET("/:id/notes/types", h.GetNoteTypes)
	spaces.GET("/:id/notes/:type", h.GetNotesByType)
	spaces.POST("/notes/search/text", h.SearchNoteByText)
	spaces.DELETE("/:space_id/notes/:note_id/delete", h.DeleteNote)

	return e, nil
}
