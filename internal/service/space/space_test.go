package space

import (
	"context"
	"errors"
	"testing"
	"time"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"
	"webserver/internal/service/space/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSpaceByID(t *testing.T) {
	type fields struct {
		repo   *mocks.Mockrepo
		cache  *mocks.MockspaceCache
		worker *mocks.MockdbWorker
	}

	type test struct {
		name       string
		spaceID    uuid.UUID
		space      model.Space
		setupMocks func(mocks *fields)
		err        error
	}

	space := model.Space{
		ID:       uuid.New(),
		Name:     "test",
		Created:  time.Now(),
		Creator:  123,
		Personal: true,
	}

	tests := []test{
		{
			name:    "positive case: from cache",
			spaceID: uuid.New(),
			space:   space,
			setupMocks: func(mocks *fields) {
				t.Helper()
				mocks.cache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(space, nil)
			},
			err: nil,
		},
		{
			name:    "positive case: from db",
			spaceID: uuid.New(),
			space:   space,
			setupMocks: func(mocks *fields) {
				t.Helper()
				mocks.cache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, api_errors.ErrSpaceNotExists)
				mocks.repo.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(space, nil)
			},
			err: nil,
		},
		{
			name:    "error case: cache error",
			spaceID: uuid.New(),
			space:   space,
			setupMocks: func(mocks *fields) {
				t.Helper()
				mocks.cache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, errors.New("cache error"))
			},
			err: errors.New("cache error"),
		},
		{
			name:    "error case: db error",
			spaceID: uuid.New(),
			space:   space,
			setupMocks: func(mocks *fields) {
				t.Helper()
				mocks.cache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, api_errors.ErrSpaceNotExists)
				mocks.repo.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, api_errors.ErrSpaceNotExists)
			},
			err: api_errors.ErrSpaceNotExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv := createTestSpace(t, repo, cache, worker)

			tt.setupMocks(&fields{repo: repo, cache: cache, worker: worker})

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
	type fields struct {
		repo   *mocks.Mockrepo
		cache  *mocks.MockspaceCache
		worker *mocks.MockdbWorker
	}

	type test struct {
		name       string
		userID     int64
		spaceID    uuid.UUID
		exists     bool
		err        error
		setupMocks func(mocks *fields)
	}

	tests := []test{
		{
			name:    "positive case",
			userID:  123,
			spaceID: uuid.New(),
			exists:  true,
			setupMocks: func(mocks *fields) {
				t.Helper()
				mocks.repo.EXPECT().CheckParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
			},
			err: nil,
		},
		{
			name:    "error case: db error",
			userID:  123,
			spaceID: uuid.New(),
			exists:  false,
			err:     errors.New("db error"),
			setupMocks: func(mocks *fields) {
				t.Helper()
				mocks.repo.EXPECT().CheckParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv := createTestSpace(t, repo, cache, worker)

			tt.setupMocks(&fields{repo: repo, cache: cache, worker: worker})

			exists, err := spaceSrv.IsUserInSpace(context.Background(), tt.userID, tt.spaceID)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.exists, exists)
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
			spaceSrv := createTestSpace(t, repo, cache, worker)

			if tt.err == nil {
				worker.EXPECT().CreateSpace(gomock.Any(), gomock.Any()).Return(nil)
			} else {
				worker.EXPECT().CreateSpace(gomock.Any(), gomock.Any()).Return(tt.err)
			}

			err := spaceSrv.CreateSpace(context.Background(), tt.req)
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
