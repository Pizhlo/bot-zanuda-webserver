package user

import (
	"context"
	"errors"
	"testing"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"
	"webserver/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type test struct {
		name  string
		repo  userRepo
		cache userCache
		want  *User
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
			want:  &User{repo: repo, cache: cache},
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
	type test struct {
		name         string
		tgID         int64
		methodErrors map[string]error
		err          error
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo, cache := createMockServices(ctrl)

	tests := []test{
		{
			name: "positive case",
			tgID: 123,
			err:  nil,
		},
		{
			name: "error case: cache error",
			tgID: 123,
			methodErrors: map[string]error{
				"cache": errors.New("cache error"),
			},
			err: errors.New("cache error"),
		},
		{
			name: "error case: repo error",
			tgID: 123,
			methodErrors: map[string]error{
				"repo": errors.New("repo error"),
			},
			err: errors.New("repo error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.methodErrors != nil {
				if err, ok := tt.methodErrors["cache"]; ok {
					cache.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{}, err)
				}

				if err, ok := tt.methodErrors["repo"]; ok {
					cache.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{}, api_errors.ErrUnknownUser)
					repo.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{}, err)
				}
			} else {
				cache.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
			}

			userSrv := createTestUserService(t, repo, cache)

			err := userSrv.CheckUser(context.Background(), tt.tgID)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func createTestUserService(t *testing.T, repo *mocks.MockuserRepo, cache *mocks.MockuserCache) *User {
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
