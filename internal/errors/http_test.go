package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPError(t *testing.T) {
	expected := &HTTPError{
		Code:       http.StatusBadRequest,
		Message:    "bad request",
		InnerError: errors.New("api error"),
	}

	actual := NewHTTPError(http.StatusBadRequest, "bad request", errors.New("api error"))

	assert.Equal(t, expected, actual)
}

func TestError(t *testing.T) {
	err := &HTTPError{
		Code:       http.StatusBadRequest,
		Message:    "bad request",
		InnerError: errors.New("api error"),
	}

	msg := err.Error()

	assert.Equal(t, "bad request", msg)
}
