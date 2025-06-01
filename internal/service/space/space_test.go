package space

import (
	"context"
	"errors"
	"testing"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"
	"webserver/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSpaceByID(t *testing.T) {
	type test struct {
		name         string
		spaceID      uuid.UUID
		space        model.Space
		methodErrors map[string]error
		err          error
	}

	tests := []test{
		{
			name:    "positive case: from cache",
			spaceID: uuid.New(),
			space: model.Space{
				ID:       uuid.New(),
				Name:     "test",
				Created:  1236788,
				Creator:  123,
				Personal: true,
			},
			err: nil,
		},
		{
			name:    "positive case: from db",
			spaceID: uuid.New(),
			space: model.Space{
				ID:       uuid.New(),
				Name:     "test",
				Created:  1236788,
				Creator:  123,
				Personal: true,
			},
			methodErrors: map[string]error{
				"cache": api_errors.ErrSpaceNotExists,
			},
			err: nil,
		},
		{
			name:    "error case: cache error",
			spaceID: uuid.New(),
			space: model.Space{
				ID:       uuid.New(),
				Name:     "test",
				Created:  1236788,
				Creator:  123,
				Personal: true,
			},
			methodErrors: map[string]error{
				"cache": errors.New("cache error"),
			},
			err: errors.New("cache error"),
		},
		{
			name:    "error case: db error",
			spaceID: uuid.New(),
			space: model.Space{
				ID:       uuid.New(),
				Name:     "test",
				Created:  1236788,
				Creator:  123,
				Personal: true,
			},
			methodErrors: map[string]error{
				"db": api_errors.ErrSpaceNotExists,
			},
			err: api_errors.ErrSpaceNotExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv, err := New(repo, cache, worker)
			require.NoError(t, err)

			// positive case
			if tt.methodErrors == nil {
				cache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(tt.space, nil)
			}

			if err, ok := tt.methodErrors["cache"]; ok {
				cache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, err)

				if tt.err == nil {
					repo.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(tt.space, nil)
				}
			}

			if err, ok := tt.methodErrors["db"]; ok {
				cache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, err)
				repo.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, err)
			}

			space, err := spaceSrv.GetSpaceByID(context.Background(), tt.spaceID)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.space, space)
			}
		})
	}
}

func TestIsUserInSpace(t *testing.T) {
	type test struct {
		name    string
		userID  int64
		spaceID uuid.UUID
		err     error
	}

	tests := []test{
		{
			name:    "positive case",
			userID:  123,
			spaceID: uuid.New(),
			err:     nil,
		},
		{
			name:    "error case: db error",
			userID:  123,
			spaceID: uuid.New(),
			err:     errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv, err := New(repo, cache, worker)
			require.NoError(t, err)

			if tt.err == nil {
				repo.EXPECT().CheckParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			} else {
				repo.EXPECT().CheckParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(tt.err)
			}

			err = spaceSrv.IsUserInSpace(context.Background(), tt.userID, tt.spaceID)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCreateSpace(t *testing.T) {
	type test struct {
		name string
		req  rabbit.CreateSpaceRequest
		err  error
	}

	tests := []test{
		{
			name: "positive case",
			req: rabbit.CreateSpaceRequest{
				ID:        uuid.New(),
				Name:      "test",
				Created:   1236788,
				UserID:    123,
				Operation: rabbit.CreateOp,
			},
			err: nil,
		},
		{
			name: "error case: db error",
			req: rabbit.CreateSpaceRequest{
				ID:        uuid.New(),
				Name:      "test",
				Created:   1236788,
				UserID:    123,
				Operation: rabbit.CreateOp,
			},
			err: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv, err := New(repo, cache, worker)
			require.NoError(t, err)

			if tt.err == nil {
				worker.EXPECT().CreateSpace(gomock.Any(), gomock.Any()).Return(nil)
			} else {
				worker.EXPECT().CreateSpace(gomock.Any(), gomock.Any()).Return(tt.err)
			}

			err = spaceSrv.CreateSpace(context.Background(), tt.req)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func createMockServices(ctrl *gomock.Controller) (*mocks.Mockrepo, *mocks.MockspaceCache, *mocks.MockdbWorker) {
	repo := mocks.NewMockrepo(ctrl)
	cache := mocks.NewMockspaceCache(ctrl)
	worker := mocks.NewMockdbWorker(ctrl)
	return repo, cache, worker
}
