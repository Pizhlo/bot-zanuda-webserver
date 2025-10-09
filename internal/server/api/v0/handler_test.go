package v0

import (
	"errors"
	"testing"

	"github.com/ex-rate/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type test struct {
		name      string
		space     spaceService
		user      userService
		auth      authService
		logger    *logger.Logger
		version   string
		buildDate string
		gitCommit string
		err       error
		want      *Handler
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger, err := logger.New(logger.Config{
		Level:  logger.DebugLevel,
		Output: logger.ConsoleOutput,
	})
	require.NoError(t, err)

	handlerLogger := logger.WithService("handler")

	spaceSrvMock, userSrvMock, authSrvMock := createMockServices(t, ctrl)

	tests := []test{
		{
			name:      "success",
			space:     spaceSrvMock,
			user:      userSrvMock,
			auth:      authSrvMock,
			logger:    handlerLogger,
			version:   "1.0.0",
			buildDate: "2021-01-01",
			gitCommit: "1234567890",
			err:       nil,
			want:      &Handler{space: spaceSrvMock, user: userSrvMock, auth: authSrvMock, logger: handlerLogger, version: "1.0.0", buildDate: "2021-01-01", gitCommit: "1234567890"},
		},
		{
			name:      "space is nil",
			space:     nil,
			user:      userSrvMock,
			auth:      authSrvMock,
			logger:    handlerLogger,
			version:   "1.0.0",
			buildDate: "2021-01-01",
			gitCommit: "1234567890",
			err:       errors.New("space is nil"),
			want:      nil,
		},
		{
			name:      "user is nil",
			space:     spaceSrvMock,
			user:      nil,
			auth:      authSrvMock,
			logger:    handlerLogger,
			version:   "1.0.0",
			buildDate: "2021-01-01",
			gitCommit: "1234567890",
			err:       errors.New("user is nil"),
			want:      nil,
		},
		{
			name:      "auth is nil",
			space:     spaceSrvMock,
			user:      userSrvMock,
			auth:      nil,
			logger:    handlerLogger,
			version:   "1.0.0",
			buildDate: "2021-01-01",
			gitCommit: "1234567890",
			err:       errors.New("auth is nil"),
			want:      nil,
		},
		{
			name:      "logger is nil",
			space:     spaceSrvMock,
			user:      userSrvMock,
			auth:      authSrvMock,
			logger:    nil,
			version:   "1.0.0",
			buildDate: "2021-01-01",
			gitCommit: "1234567890",
			err:       errors.New("logger is nil"),
			want:      nil,
		},
		{
			name:      "version is nil",
			space:     spaceSrvMock,
			user:      userSrvMock,
			auth:      authSrvMock,
			logger:    handlerLogger,
			version:   "",
			buildDate: "2021-01-01",
			gitCommit: "1234567890",
			err:       errors.New("version is nil"),
			want:      nil,
		},
		{
			name:      "buildDate is nil",
			space:     spaceSrvMock,
			user:      userSrvMock,
			auth:      authSrvMock,
			logger:    handlerLogger,
			version:   "1.0.0",
			buildDate: "",
			gitCommit: "1234567890",
			err:       errors.New("buildDate is nil"),
			want:      nil,
		},
		{
			name:      "gitCommit is nil",
			space:     spaceSrvMock,
			user:      userSrvMock,
			auth:      authSrvMock,
			logger:    handlerLogger,
			version:   "1.0.0",
			buildDate: "2021-01-01",
			gitCommit: "",
			err:       errors.New("gitCommit is nil"),
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := New(
				WithSpaceService(tt.space),
				WithUserService(tt.user),
				WithAuthService(tt.auth),
				WithLogger(tt.logger),
				WithVersion(tt.version),
				WithBuildDate(tt.buildDate),
				WithGitCommit(tt.gitCommit),
			)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
				assert.Nil(t, handler)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, handler)
			}
		})
	}
}
