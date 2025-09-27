package space

import (
	"errors"
	"testing"
	"webserver/internal/service/space/mocks"

	"github.com/ex-rate/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type test struct {
		name   string
		repo   repo
		cache  spaceCache
		worker dbWorker
		logger *logger.Logger
		want   *Service
		err    error
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo, cache, worker := createMockServices(ctrl)

	logger, err := logger.New(logger.Config{
		Level:  logger.DebugLevel,
		Output: logger.ConsoleOutput,
	})
	require.NoError(t, err)

	spaceLogger := logger.WithService("space")

	tests := []test{
		{
			name:   "positive case",
			repo:   repo,
			cache:  cache,
			worker: worker,
			logger: spaceLogger,
			want:   &Service{repo: repo, cache: cache, worker: worker, logger: spaceLogger},
			err:    nil,
		},
		{
			name:   "error case: repo is nil",
			repo:   nil,
			cache:  cache,
			worker: worker,
			logger: spaceLogger,
			err:    errors.New("repo is nil"),
		},
		{
			name:   "error case: cache is nil",
			repo:   repo,
			cache:  nil,
			worker: worker,
			logger: spaceLogger,
			err:    errors.New("cache is nil"),
		},
		{
			name:   "error case: worker is nil",
			repo:   repo,
			cache:  cache,
			worker: nil,
			logger: spaceLogger,
			err:    errors.New("worker is nil"),
		},
		{
			name:   "error case: logger is nil",
			repo:   repo,
			cache:  cache,
			worker: worker,
			logger: nil,
			err:    errors.New("logger is nil"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spaceSrv, err := New(
				WithRepo(tt.repo),
				WithCache(tt.cache),
				WithWorker(tt.worker),
				WithLogger(tt.logger),
			)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
				assert.Nil(t, spaceSrv)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, spaceSrv)
			}
		})
	}
}

func createTestSpaceSrv(t *testing.T, repo *mocks.Mockrepo, cache *mocks.MockspaceCache, worker *mocks.MockdbWorker) *Service {
	logger, err := logger.New(logger.Config{
		Level:  logger.DebugLevel,
		Output: logger.ConsoleOutput,
	})
	require.NoError(t, err)

	spaceLogger := logger.WithService("space")

	spaceSrv, err := New(
		WithRepo(repo),
		WithCache(cache),
		WithWorker(worker),
		WithLogger(spaceLogger),
	)
	require.NoError(t, err)

	return spaceSrv
}
