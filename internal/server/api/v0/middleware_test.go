package v0

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"
	"webserver/internal/service/space"
	"webserver/internal/service/user"
	"webserver/mocks"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/undefinedlabs/go-mpatch"
)

// тест для проверки middleware
func TestValidateNoteRequest_CreateNote(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	type test struct {
		name         string
		req          rabbit.CreateNoteRequest
		expectedNote rabbit.CreateNoteRequest
		dbErr        bool // должна ли база вернуть ошибку
		// ошибки разных репозиториев
		err              error            // ошибки валидации и т.п.
		methodErrors     map[string]error // название метода : ошибка
		expectedCode     int
		expectedResponse map[string]string
	}

	generatedID := uuid.New()

	uuidPatch, err := mpatch.PatchMethod(uuid.New, func() uuid.UUID { return generatedID })
	require.NoError(t, err)

	defer uuidPatch.Unpatch()

	wayback := time.Now()
	timePatch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer timePatch.Unpatch()

	tests := []test{
		{
			name: "create note",
			req: rabbit.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			expectedNote: rabbit.CreateNoteRequest{
				ID:        uuid.New(),
				UserID:    1,
				Text:      "new note",
				SpaceID:   uuid.New(),
				Type:      model.TextNoteType,
				Created:   time.Now().Unix(),
				Operation: rabbit.CreateOp,
			},
			expectedCode: http.StatusAccepted,
			expectedResponse: map[string]string{
				"request_id": uuid.New().String(),
			},
		},
		{
			name:         "db err: unknown user",
			dbErr:        true,
			err:          api_errors.ErrUnknownUser,
			methodErrors: map[string]error{"CheckUser": api_errors.ErrUnknownUser},
			req: rabbit.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: map[string]string{"bad request": "unknown user"},
		},
		{
			name:         "db err: space not exists",
			dbErr:        true,
			err:          api_errors.ErrSpaceNotExists,
			methodErrors: map[string]error{"GetSpaceByID": api_errors.ErrSpaceNotExists},
			req: rabbit.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: map[string]string{"bad request": "space does not exist"},
		},
		{
			name:         "db err: space belongs another user",
			dbErr:        true,
			err:          api_errors.ErrSpaceNotBelongsUser,
			methodErrors: map[string]error{"IsUserInSpace": api_errors.ErrSpaceNotBelongsUser},
			req: rabbit.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: map[string]string{"bad request": "space not belongs to user"},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceRepo := mocks.NewMockrepo(ctrl)
	spaceCache := mocks.NewMockspaceCache(ctrl)
	saver := mocks.NewMockdbWorker(ctrl)

	spaceSrv := space.New(spaceRepo, spaceCache, saver)

	userRepo := mocks.NewMockuserRepo(ctrl)
	userCache := mocks.NewMockuserCache(ctrl)

	userSrv := user.New(userRepo, userCache)

	handler := New(spaceSrv, userSrv, nil)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbErr {
				if err, ok := tt.methodErrors["CheckUser"]; ok {
					userRepo.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{}, err)
					userCache.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{}, err)
				}

				if err, ok := tt.methodErrors["GetSpaceByID"]; ok {
					userCache.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{ID: 1}, nil)

					spaceCache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, err)
					spaceRepo.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, err)
				}

				if err, ok := tt.methodErrors["IsUserInSpace"]; ok {
					userCache.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{ID: 1}, nil)
					spaceCache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, nil)

					spaceRepo.EXPECT().CheckParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(err)
				}
			}

			// positive case
			if tt.err == nil {
				userCache.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{
					ID: 1,
				}, nil)

				spaceRepo.EXPECT().CheckParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

				spaceCache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{
					ID: generatedID,
				}, nil)

				saver.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(nil).Do(func(ctx any, actualReq rabbit.CreateNoteRequest) {
					assert.Equal(t, tt.expectedNote, actualReq, "requests not equal")
				})
			}

			bodyJSON, err := json.Marshal(tt.req)
			require.NoError(t, err)

			resp := testRequest(t, ts, http.MethodPost, "/api/v0/spaces/notes/create", bytes.NewReader(bodyJSON))
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedResponse != nil {
				var result map[string]string
				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedResponse, result, "result IDs not equal")
			}
		})
	}
}

func TestValidateNoteRequest_UpdateNote(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	type test struct {
		name         string
		req          rabbit.UpdateNoteRequest
		expectedNote rabbit.UpdateNoteRequest
		dbNote       model.GetNote // что возвращает база
		dbErr        bool          // должна ли база вернуть ошибку
		// ошибки разных репозиториев
		err              error            // ошибки валидации и т.п.
		methodErrors     map[string]error // название метода : ошибка
		expectedCode     int
		expectedResponse map[string]string
	}

	generatedID := uuid.New()

	uuidPatch, err := mpatch.PatchMethod(uuid.New, func() uuid.UUID { return generatedID })
	require.NoError(t, err)

	defer uuidPatch.Unpatch()

	wayback := time.Now()
	timePatch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer timePatch.Unpatch()

	tests := []test{
		{
			name: "update note",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: generatedID,
				NoteID:  generatedID,
			},
			dbNote: model.GetNote{
				UserID:  1,
				Text:    "new note",
				SpaceID: generatedID,
				ID:      generatedID,
				Type:    model.TextNoteType,
			},
			expectedNote: rabbit.UpdateNoteRequest{
				ID:        generatedID,
				UserID:    1,
				Text:      "new note",
				SpaceID:   generatedID,
				NoteID:    generatedID,
				Created:   time.Now().Unix(),
				Operation: rabbit.UpdateOp,
			},
			expectedCode: http.StatusAccepted,
			expectedResponse: map[string]string{
				"request_id": uuid.New().String(),
			},
		},
		{
			name:         "db err: unknown user",
			dbErr:        true,
			err:          api_errors.ErrUnknownUser,
			methodErrors: map[string]error{"CheckUser": api_errors.ErrUnknownUser},
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: map[string]string{"bad request": "unknown user"},
		},
		{
			name:         "db err: space not exists",
			dbErr:        true,
			err:          api_errors.ErrSpaceNotExists,
			methodErrors: map[string]error{"GetSpaceByID": api_errors.ErrSpaceNotExists},
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: map[string]string{"bad request": "space does not exist"},
		},
		{
			name:         "db err: space belongs another user",
			dbErr:        true,
			err:          api_errors.ErrSpaceNotBelongsUser,
			methodErrors: map[string]error{"IsUserInSpace": api_errors.ErrSpaceNotBelongsUser},
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: map[string]string{"bad request": "space not belongs to user"},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceRepo := mocks.NewMockrepo(ctrl)
	spaceCache := mocks.NewMockspaceCache(ctrl)
	saver := mocks.NewMockdbWorker(ctrl)

	spaceSrv := space.New(spaceRepo, spaceCache, saver)

	userRepo := mocks.NewMockuserRepo(ctrl)
	userCache := mocks.NewMockuserCache(ctrl)

	userSrv := user.New(userRepo, userCache)

	handler := New(spaceSrv, userSrv, nil)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbErr {
				if err, ok := tt.methodErrors["CheckUser"]; ok {
					userRepo.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{}, err)
					userCache.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{}, err)
				}

				if err, ok := tt.methodErrors["GetSpaceByID"]; ok {
					userCache.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{ID: 1}, nil)

					spaceCache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, err)
					spaceRepo.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, err)
				}

				if err, ok := tt.methodErrors["IsUserInSpace"]; ok {
					userCache.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{ID: 1}, nil)
					spaceCache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, nil)

					spaceRepo.EXPECT().CheckParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(err)
				}
			}

			// positive case
			if tt.err == nil {
				userCache.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{
					ID: 1,
				}, nil)

				spaceRepo.EXPECT().GetNoteByID(gomock.Any(), gomock.Any()).Return(tt.dbNote, nil)

				spaceRepo.EXPECT().CheckParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

				spaceCache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{
					ID: generatedID,
				}, nil)

				saver.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(nil).Do(func(ctx any, actualReq rabbit.UpdateNoteRequest) {
					assert.Equal(t, tt.expectedNote, actualReq, "requests not equal")
				})
			}

			bodyJSON, err := json.Marshal(tt.req)
			require.NoError(t, err)

			resp := testRequest(t, ts, http.MethodPatch, "/api/v0/spaces/notes/update", bytes.NewReader(bodyJSON))
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedResponse != nil {
				var result map[string]string
				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedResponse, result, "result IDs not equal")
			}
		})
	}
}
