package v0

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"webserver/internal/server/api/v0/mocks"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, token string, body io.Reader) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	req.Close = true
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("User-Agent", "PostmanRuntime/7.32.3")
	require.NoError(t, err)

	if token != "" {
		req.Header.Set("Authorization", token)
	}

	ts.Client()

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	return resp
}

func runTestServer(t *testing.T, h *Handler) (*echo.Echo, error) {
	t.Helper()

	e := echo.New()

	api := e.Group("api/")

	// v0
	apiv0 := api.Group("v0/")

	apiv0.GET("health", h.Health)

	spaces := apiv0.Group("spaces")

	// spaces
	spaces.POST("/create", h.CreateSpace, h.Auth, h.WrapNetHTTP)                        // создать пространство
	spaces.POST("/:space_id/participants/add", h.AddParticipant, h.Auth, h.WrapNetHTTP) // добавить участника в пространство

	// notes
	spaces.GET("/:space_id/notes", h.NotesBySpaceID, h.WrapNetHTTP)

	// создание, обновление, удаление
	spaces.POST("/notes/create", h.CreateNote, h.WrapNetHTTP)
	spaces.PATCH("/notes/update", h.UpdateNote, h.WrapNetHTTP)
	spaces.DELETE("/:space_id/notes/:note_id/delete", h.DeleteNote, h.WrapNetHTTP)
	spaces.DELETE("/:space_id/notes/delete_all", h.DeleteAllNotes, h.WrapNetHTTP) // удалить все заметки

	// типы заметок
	spaces.GET("/:space_id/notes/types", h.GetNoteTypes, h.WrapNetHTTP)   // получить, какие есть типы заметок
	spaces.GET("/:space_id/notes/:type", h.GetNotesByType, h.WrapNetHTTP) // получить все заметки одного типа

	// поиск
	spaces.POST("/notes/search/text", h.SearchNoteByText, h.WrapNetHTTP) // по тексту

	return e, nil
}
func runTestServerWithMiddleware(t *testing.T, h *Handler) (*echo.Echo, error) {
	t.Helper()

	e := echo.New()

	api := e.Group("api/")

	// v0
	apiv0 := api.Group("v0/")

	apiv0.GET("health", h.Health)

	spaces := apiv0.Group("spaces")

	// spaces
	spaces.POST("/create", h.CreateSpace, h.Auth)                        // создать пространство
	spaces.POST("/:space_id/participants/add", h.AddParticipant, h.Auth) // добавить участника в пространство

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

func createMockServices(t *testing.T, ctrl *gomock.Controller) (*mocks.MockspaceService, *mocks.MockuserService, *mocks.MockauthService) {
	t.Helper()

	spaceSrvMock := mocks.NewMockspaceService(ctrl)
	userSrvMock := mocks.NewMockuserService(ctrl)
	authSrvMock := mocks.NewMockauthService(ctrl)

	return spaceSrvMock, userSrvMock, authSrvMock
}

func createTestHandler(t *testing.T, ctrl *gomock.Controller) *Handler {
	t.Helper()

	spaceSrvMock, userSrvMock, authSrvMock := createMockServices(t, ctrl)

	h, err := New(WithSpaceService(spaceSrvMock), WithUserService(userSrvMock), WithAuthService(authSrvMock))
	require.NoError(t, err)

	return h
}

func generateToken(t *testing.T, userID, expired float64) string {
	t.Helper()

	claims := jwt.MapClaims{}

	if userID != 0 {
		claims["user_id"] = userID
	}

	if expired != 0 {
		claims["expired"] = expired
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("secret"))
	require.NoError(t, err)

	return tokenString
}

func checkResult(t *testing.T, resp *http.Response, expectedResp error) {
	t.Helper()

	var result map[string]string

	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, expectedResp.Error(), result["error"])
}

func checkRequestID(t *testing.T, resp *http.Response) {
	t.Helper()

	var respBody map[string]string
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	_, err = uuid.Parse(respBody["request_id"])
	require.NoError(t, err)
}
