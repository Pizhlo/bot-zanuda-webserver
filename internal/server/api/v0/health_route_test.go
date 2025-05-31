package v0

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceSrvMock, userSrvMock, authSrvMock := createMockServices(ctrl)

	handler, err := New(spaceSrvMock, userSrvMock, authSrvMock)
	require.NoError(t, err)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp := testRequest(t, ts, http.MethodGet, "/api/v0/health", nil)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, http.NoBody, resp.Body)
}
