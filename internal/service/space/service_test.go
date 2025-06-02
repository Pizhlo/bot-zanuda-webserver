package space

import (
	"errors"
	"testing"
	"webserver/internal/service/space/mocks"

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
		want   *space
		err    error
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo, cache, worker := createMockServices(ctrl)

	tests := []test{
		{
			name:   "positive case",
			repo:   repo,
			cache:  cache,
			worker: worker,
			want:   &space{repo: repo, cache: cache, worker: worker},
			err:    nil,
		},
		{
			name:   "error case: repo is nil",
			repo:   nil,
			cache:  cache,
			worker: worker,
			err:    errors.New("repo is nil"),
		},
		{
			name:   "error case: cache is nil",
			repo:   repo,
			cache:  nil,
			worker: worker,
			err:    errors.New("cache is nil"),
		},
		{
			name:   "error case: worker is nil",
			repo:   repo,
			cache:  cache,
			worker: nil,
			err:    errors.New("worker is nil"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spaceSrv, err := New(
				WithRepo(tt.repo),
				WithCache(tt.cache),
				WithWorker(tt.worker),
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

func createTestSpace(t *testing.T, repo *mocks.Mockrepo, cache *mocks.MockspaceCache, worker *mocks.MockdbWorker) *space {
	spaceSrv, err := New(
		WithRepo(repo),
		WithCache(cache),
		WithWorker(worker),
	)
	require.NoError(t, err)

	return spaceSrv
}
