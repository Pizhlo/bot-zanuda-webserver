package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	r, err := runTestServer()
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp := testRequest(t, ts, http.MethodGet, "/health", nil)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, http.NoBody, resp.Body)
}
