package auth

import (
	"errors"
	"testing"

	"github.com/ex-rate/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type test struct {
		name string
		opts []AuthOption
		want *Service
		err  error
	}

	secretKey := []byte("secret")

	logger, err := logger.New(logger.Config{
		Level:  logger.DebugLevel,
		Output: logger.ConsoleOutput,
	})
	require.NoError(t, err)

	authLogger := logger.WithService("auth")

	tests := []test{
		{
			name: "positive case",
			opts: []AuthOption{
				WithSecretKey(secretKey),
				WithLogger(authLogger),
			},
			want: &Service{
				secretKey: secretKey,
				logger:    authLogger,
			},
			err: nil,
		},
		{
			name: "error case: secret key is required",
			opts: []AuthOption{
				WithLogger(authLogger),
			},
			err: errors.New("secret key is required"),
		},
		{
			name: "error case: logger is required",
			opts: []AuthOption{
				WithSecretKey(secretKey),
			},
			err: errors.New("logger is nil"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := New(tt.opts...)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
				assert.Nil(t, auth)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, auth)
			}
		})
	}
}

func createTestAuthService(t *testing.T, secretKey []byte) *Service {
	logger, err := logger.New(logger.Config{
		Level:  logger.DebugLevel,
		Output: logger.ConsoleOutput,
	})
	require.NoError(t, err)

	authLogger := logger.WithService("auth")

	auth, err := New(WithSecretKey(secretKey), WithLogger(authLogger))
	require.NoError(t, err)
	return auth
}
