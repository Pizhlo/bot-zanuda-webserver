package user

import (
	"context"
	"errors"
	"testing"
	"webserver/internal/service/user/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type test struct {
		name  string
		repo  userRepo
		cache userCache
		want  *Service
		err   error
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo, cache := createMockServices(ctrl)

	tests := []test{
		{
			name:  "positive case",
			repo:  repo,
			cache: cache,
			want:  &Service{repo: repo, cache: cache},
			err:   nil,
		},
		{
			name:  "error case: repo is nil",
			repo:  nil,
			cache: cache,
			err:   errors.New("repo is nil"),
		},
		{
			name:  "error case: cache is nil",
			repo:  repo,
			cache: nil,
			err:   errors.New("cache is nil"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userSrv, err := New(
				WithRepo(tt.repo),
				WithCache(tt.cache),
			)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
				assert.Nil(t, userSrv)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, userSrv)
			}
		})
	}
}

func TestCheckUser(t *testing.T) {
	type fields struct {
		repo  *mocks.MockuserRepo
		cache *mocks.MockuserCache
	}

	type test struct {
		name       string
		tgID       int64
		err        error
		exists     bool
		setupMocks func(mocks *fields)
	}

	tests := []test{
		{
			name:   "positive case: exists in cache",
			tgID:   123,
			err:    nil,
			exists: true,
			setupMocks: func(mocks *fields) {
				t.Helper()
				mocks.cache.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
			},
		},
		{
			name:   "positive case: exists in repo",
			tgID:   123,
			err:    nil,
			exists: true,
			setupMocks: func(mocks *fields) {
				t.Helper()
				mocks.cache.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(false, nil)
				mocks.repo.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
			},
		},
		{
			name:   "error case: cache error",
			tgID:   123,
			err:    errors.New("cache error"),
			exists: false,
			setupMocks: func(mocks *fields) {
				t.Helper()
				mocks.cache.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(false, errors.New("cache error"))
			},
		},
		{
			name:   "error case: repo error",
			tgID:   123,
			exists: false,
			err:    errors.New("repo error"),
			setupMocks: func(mocks *fields) {
				t.Helper()
				mocks.cache.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(false, nil)
				mocks.repo.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(false, errors.New("repo error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache := createMockServices(ctrl)

			tt.setupMocks(&fields{repo: repo, cache: cache})

			userSrv := createTestUserService(t, repo, cache)

			exists, err := userSrv.CheckUser(context.Background(), tt.tgID)
			assert.Equal(t, tt.exists, exists)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func createTestUserService(t *testing.T, repo *mocks.MockuserRepo, cache *mocks.MockuserCache) *Service {
	userSrv, err := New(
		WithRepo(repo),
		WithCache(cache),
	)
	require.NoError(t, err)

	return userSrv
}

func createMockServices(ctrl *gomock.Controller) (*mocks.MockuserRepo, *mocks.MockuserCache) {
	repo := mocks.NewMockuserRepo(ctrl)
	cache := mocks.NewMockuserCache(ctrl)

	return repo, cache
}
