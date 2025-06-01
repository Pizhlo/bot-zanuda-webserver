package v0

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"webserver/mocks"

	"github.com/golang/mock/gomock"
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

func runTestServer(h *handler) (*echo.Echo, error) {
	e := echo.New()

	api := e.Group("api/")

	// v0
	apiv0 := api.Group("v0/")

	apiv0.GET("health", h.Health)

	spaces := apiv0.Group("spaces")

	// notes
	spaces.GET("/:space_id/notes", h.NotesBySpaceID)

	// создание, обновление, удаление
	spaces.POST("/notes/create", h.CreateNote)
	spaces.PATCH("/notes/update", h.UpdateNote)
	spaces.DELETE("/:space_id/notes/:note_id/delete", h.DeleteNote)
	spaces.DELETE("/:space_id/notes/delete_all", h.DeleteAllNotes) // удалить все заметки

	// типы заметок
	spaces.GET("/:space_id/notes/types", h.GetNoteTypes)   // получить, какие есть типы заметок
	spaces.GET("/:space_id/notes/:type", h.GetNotesByType) // получить все заметки одного типа

	// поиск
	spaces.POST("/notes/search/text", h.SearchNoteByText) // по тексту

	return e, nil
}
func runTestServerWithMiddleware(h *handler) (*echo.Echo, error) {
	e := echo.New()

	api := e.Group("api/")

	// v0
	apiv0 := api.Group("v0/")

	apiv0.GET("health", h.Health)

	spaces := apiv0.Group("spaces")

	// notes
	spaces.GET("/:space_id/notes", h.NotesBySpaceID)

	// создание, обновление, удаление
	spaces.POST("/notes/create", h.CreateNote, h.ValidateNoteRequest)
	spaces.PATCH("/notes/update", h.UpdateNote, h.ValidateNoteRequest)
	spaces.DELETE("/:space_id/notes/:note_id/delete", h.DeleteNote)
	spaces.DELETE("/:space_id/notes/delete_all", h.DeleteAllNotes) // удалить все заметки

	// типы заметок
	spaces.GET("/:space_id/notes/types", h.GetNoteTypes)   // получить, какие есть типы заметок
	spaces.GET("/:space_id/notes/:type", h.GetNotesByType) // получить все заметки одного типа

	// поиск
	spaces.POST("/notes/search/text", h.SearchNoteByText) // по тексту

	return e, nil
}

func createMockServices(ctrl *gomock.Controller) (*mocks.MockspaceService, *mocks.MockuserService, *mocks.MockauthService) {
	spaceSrvMock := mocks.NewMockspaceService(ctrl)
	userSrvMock := mocks.NewMockuserService(ctrl)
	authSrvMock := mocks.NewMockauthService(ctrl)

	return spaceSrvMock, userSrvMock, authSrvMock
}

func createTestHandler(t *testing.T, ctrl *gomock.Controller) *handler {
	spaceSrvMock, userSrvMock, authSrvMock := createMockServices(ctrl)

	h, err := New(WithSpaceService(spaceSrvMock), WithUserService(userSrvMock), WithAuthService(authSrvMock))
	require.NoError(t, err)

	return h
}
