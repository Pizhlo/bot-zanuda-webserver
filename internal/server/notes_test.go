package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"
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

func TestCreateNote(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	type test struct {
		name             string
		req              model.CreateNoteRequest
		expectedNote     model.CreateNoteRequest
		err              bool
		expectedCode     int
		expectedResponse map[string]string
	}

	generatedID := uuid.New()

	uuidPatch, err := mpatch.PatchMethod(uuid.New, func() uuid.UUID { return generatedID })
	require.NoError(t, err)

	defer uuidPatch.Unpatch()

	tests := []test{
		{
			name: "positive test",
			req: model.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			expectedNote: model.CreateNoteRequest{
				ID:      uuid.New(),
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			expectedCode: http.StatusAccepted,
			expectedResponse: map[string]string{
				"request_id": uuid.New().String(),
			},
		},
		{
			name: "user ID not filled",
			req: model.CreateNoteRequest{
				Text:    "new note",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			err:          true,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrFieldUserNotFilled.Error(),
			},
		},
		{
			name: "text not filled",
			req: model.CreateNoteRequest{
				UserID:  1,
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			err:          true,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrFieldTextNotFilled.Error(),
			},
		},
		{
			name: "invalid space ID",
			req: model.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.Nil,
				Type:    model.TextNoteType,
			},
			err:          true,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrInvalidSpaceID.Error(),
			},
		},
		{
			name: "field type not filled",
			req: model.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
			},
			err:          true,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrFieldTypeNotFilled.Error(),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceRepo := mocks.NewMockspaceRepo(ctrl)
	spaceCache := mocks.NewMockspaceCache(ctrl)
	saver := mocks.NewMockdbWorker(ctrl)

	spaceSrv := space.New(spaceRepo, spaceCache, saver)

	userRepo := mocks.NewMockuserRepo(ctrl)
	userCache := mocks.NewMockuserCache(ctrl)

	userSrv := user.New(userRepo, userCache)

	server := New("", spaceSrv, userSrv)

	r, err := runTestServer(server)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCache.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{
				ID: 1,
			}, nil)

			spaceRepo.EXPECT().CheckParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			spaceCache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{
				ID: generatedID,
			}, nil)

			if !tt.err {
				saver.EXPECT().CreateNote(gomock.Any()).Return(nil).Do(func(actualReq model.CreateNoteRequest) {
					assert.Equal(t, tt.expectedNote, actualReq, "requests not equal")
				})
			}

			bodyJSON, err := json.Marshal(tt.req)
			require.NoError(t, err)

			resp := testRequest(t, ts, http.MethodPost, "/spaces/notes/create", bytes.NewReader(bodyJSON))
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

func TestUpdateNote(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	type test struct {
		name             string
		req              model.UpdateNoteRequest
		expectedNote     model.UpdateNoteRequest
		dbNote           model.GetNote // что возвращает база при вызове GetNote
		err              bool
		getNote          bool // нужно ли вызывать getNote
		expectedCode     int
		expectedResponse map[string]string
		methodErrors     map[string]error // название метода : ошибка
	}

	generatedID := uuid.New()
	newID := uuid.New() // для случая, когда айди должен отличаться (заметка не принадлежит пространству)

	tests := []test{
		{
			name: "positive test",
			req: model.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: generatedID,
			},
			getNote: true,
			dbNote: model.GetNote{
				UserID:  1,
				ID:      generatedID,
				Text:    "new note",
				SpaceID: generatedID,
				Type:    model.TextNoteType,
				Created: time.Now(),
			},
			expectedNote: model.UpdateNoteRequest{
				UserID:  1,
				ID:      generatedID,
				Text:    "new note",
				SpaceID: generatedID,
			},
			expectedCode: http.StatusAccepted,
			expectedResponse: map[string]string{
				"request_id": uuid.New().String(),
			},
		},
		{
			name: "user ID not filled",
			req: model.UpdateNoteRequest{
				Text:    "new note",
				SpaceID: uuid.New(),
			},
			err:          true,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrFieldUserNotFilled.Error(),
			},
		},
		{
			name: "text not filled",
			req: model.UpdateNoteRequest{
				UserID:  1,
				SpaceID: uuid.New(),
			},
			err:          true,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrFieldTextNotFilled.Error(),
			},
		},
		{
			name: "invalid space ID",
			req: model.UpdateNoteRequest{
				UserID:  1,
				Text:    "invalid space ID",
				SpaceID: uuid.Nil,
				NoteID:  generatedID,
			},
			err:          true,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrInvalidSpaceID.Error(),
			},
		},
		{
			name: "note not belongs space",
			req: model.UpdateNoteRequest{
				UserID:  1,
				Text:    "note not belongs space REQ",
				SpaceID: generatedID,
				NoteID:  generatedID,
			},
			dbNote: model.GetNote{
				UserID:  1,
				ID:      generatedID,
				Text:    "note not belongs space DB",
				SpaceID: newID,
				Type:    model.TextNoteType,
				Created: time.Now(),
			},
			err:          true,
			getNote:      true,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": api_errors.ErrNoteNotBelongsSpace.Error(),
			},
		},
		{
			name: "note type is not text",
			req: model.UpdateNoteRequest{
				UserID:  1,
				Text:    "note type is not text REQ",
				SpaceID: generatedID,
				NoteID:  generatedID,
			},
			err:     true,
			getNote: true,
			dbNote: model.GetNote{
				UserID:  1,
				ID:      generatedID,
				Text:    "note type is not text DB",
				SpaceID: generatedID,
				Type:    model.PhotoNoteType,
				Created: time.Now(),
			},
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrUpdateNotTextNote.Error(),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceRepo := mocks.NewMockspaceRepo(ctrl)
	spaceCache := mocks.NewMockspaceCache(ctrl)
	saver := mocks.NewMockdbWorker(ctrl)

	spaceSrv := space.New(spaceRepo, spaceCache, saver)

	userRepo := mocks.NewMockuserRepo(ctrl)
	userCache := mocks.NewMockuserCache(ctrl)

	userSrv := user.New(userRepo, userCache)

	server := New("", spaceSrv, userSrv)

	r, err := runTestServer(server)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCache.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(model.User{
				ID: 1,
			}, nil)

			spaceRepo.EXPECT().CheckParticipant(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			spaceCache.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{
				ID: generatedID,
			}, nil)

			if err, ok := tt.methodErrors["CheckIfNoteExistsInSpace"]; ok {
				spaceRepo.EXPECT().CheckIfNoteExistsInSpace(gomock.Any(), gomock.Any(), gomock.Any()).Return(err)
			}

			logrus.Debugf("tt.name: %s. tt.dbNote: %+v", tt.name, tt.dbNote)

			if tt.getNote {
				spaceRepo.EXPECT().GetNoteByID(gomock.Any(), gomock.Any()).Return(tt.dbNote, nil)
			}

			if !tt.err {
				uuidPatch, err := mpatch.PatchMethod(uuid.New, func() uuid.UUID { return generatedID })
				require.NoError(t, err)

				defer uuidPatch.Unpatch()

				tt.expectedResponse = map[string]string{"request_id": uuid.New().String()}

				saver.EXPECT().UpdateNote(gomock.Any()).Return(nil).Do(func(actualReq model.UpdateNoteRequest) {
					assert.Equal(t, tt.expectedNote, actualReq, "requests not equal")
				})
			}

			bodyJSON, err := json.Marshal(tt.req)
			require.NoError(t, err)

			resp := testRequest(t, ts, http.MethodPatch, "/spaces/notes/update", bytes.NewReader(bodyJSON))
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

// тест для проверки middleware
func TestValidateNoteRequest_CreateNote(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	type test struct {
		name         string
		req          model.CreateNoteRequest
		expectedNote model.CreateNoteRequest
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

	tests := []test{
		{
			name: "create note",
			req: model.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			expectedNote: model.CreateNoteRequest{
				ID:      uuid.New(),
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
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
			req: model.CreateNoteRequest{
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
			req: model.CreateNoteRequest{
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
			req: model.CreateNoteRequest{
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

	spaceRepo := mocks.NewMockspaceRepo(ctrl)
	spaceCache := mocks.NewMockspaceCache(ctrl)
	saver := mocks.NewMockdbWorker(ctrl)

	spaceSrv := space.New(spaceRepo, spaceCache, saver)

	userRepo := mocks.NewMockuserRepo(ctrl)
	userCache := mocks.NewMockuserCache(ctrl)

	userSrv := user.New(userRepo, userCache)

	server := New("", spaceSrv, userSrv)

	r, err := runTestServer(server)
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

				saver.EXPECT().CreateNote(gomock.Any()).Return(nil).Do(func(actualReq model.CreateNoteRequest) {
					assert.Equal(t, tt.expectedNote, actualReq, "requests not equal")
				})
			}

			bodyJSON, err := json.Marshal(tt.req)
			require.NoError(t, err)

			resp := testRequest(t, ts, http.MethodPost, "/spaces/notes/create", bytes.NewReader(bodyJSON))
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
		req          model.UpdateNoteRequest
		expectedNote model.UpdateNoteRequest
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

	tests := []test{
		{
			name: "update note",
			req: model.UpdateNoteRequest{
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
			expectedNote: model.UpdateNoteRequest{
				ID:      generatedID,
				UserID:  1,
				Text:    "new note",
				SpaceID: generatedID,
				NoteID:  generatedID,
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
			req: model.UpdateNoteRequest{
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
			req: model.UpdateNoteRequest{
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
			req: model.UpdateNoteRequest{
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

	spaceRepo := mocks.NewMockspaceRepo(ctrl)
	spaceCache := mocks.NewMockspaceCache(ctrl)
	saver := mocks.NewMockdbWorker(ctrl)

	spaceSrv := space.New(spaceRepo, spaceCache, saver)

	userRepo := mocks.NewMockuserRepo(ctrl)
	userCache := mocks.NewMockuserCache(ctrl)

	userSrv := user.New(userRepo, userCache)

	server := New("", spaceSrv, userSrv)

	r, err := runTestServer(server)
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

				saver.EXPECT().UpdateNote(gomock.Any()).Return(nil).Do(func(actualReq model.UpdateNoteRequest) {
					assert.Equal(t, tt.expectedNote, actualReq, "requests not equal")
				})
			}

			bodyJSON, err := json.Marshal(tt.req)
			require.NoError(t, err)

			resp := testRequest(t, ts, http.MethodPatch, "/spaces/notes/update", bytes.NewReader(bodyJSON))
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

func TestNotesBySpaceID_Full(t *testing.T) {
	type test struct {
		name             string
		spaceID          string
		dbErr            error // ошибка, которую возвращает база
		expectedCode     int
		expectedResponse []model.Note
		expectedErr      map[string]string
	}

	wayback := time.Date(2024, time.May, 19, 1, 2, 3, 4, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer patch.Unpatch()

	tests := []test{
		{
			name:         "positive test",
			spaceID:      "1234",
			expectedCode: http.StatusOK,
			expectedResponse: []model.Note{
				{
					ID: uuid.New(),
					User: &model.User{
						ID:       1,
						TgID:     1234,
						Username: "test user",
						PersonalSpace: model.Space{
							ID:       uuid.New(),
							Name:     "personal space for user 1234",
							Created:  time.Now(),
							Creator:  1,
							Personal: true,
						},
						Timezone: "Europe/Moscow",
					},
					Text: "test note",
					Space: &model.Space{
						ID:       uuid.New(),
						Name:     "personal space for user 1234",
						Created:  time.Now(),
						Creator:  1,
						Personal: true,
					},
					Created:  time.Now(),
					LastEdit: sql.NullTime{Valid: false},
				},
			},
		},
		{
			name:         "space does not have any notes",
			spaceID:      "1234",
			dbErr:        api_errors.ErrNoNotesFoundBySpaceID,
			expectedCode: http.StatusNoContent,
			expectedErr:  nil,
		},
		{
			name:         "space does not exist",
			spaceID:      "1234",
			dbErr:        api_errors.ErrSpaceNotExists,
			expectedCode: http.StatusNotFound,
			expectedErr:  nil,
		},
		{
			name:         "invalid param",
			spaceID:      "1234abc",
			expectedCode: http.StatusBadRequest,
			expectedErr:  map[string]string{"bad request": "invalid space id parameter: strconv.Atoi: parsing \"1234abc\": invalid syntax"},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockspaceRepo(ctrl)

	spaceSrv := space.New(repo, nil, nil)

	server := New("", spaceSrv, nil)

	r, err := runTestServer(server)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbErr != nil {
				repo.EXPECT().GetAllNotesBySpaceIDFull(gomock.Any(), gomock.Any()).Return(nil, tt.dbErr)
			} else if tt.expectedCode == http.StatusOK {
				repo.EXPECT().GetAllNotesBySpaceIDFull(gomock.Any(), gomock.Any()).Return(tt.expectedResponse, nil)
			}

			url := fmt.Sprintf("/spaces/%s/notes?full_user=true", tt.spaceID)

			resp := testRequest(t, ts, http.MethodGet, url, nil)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			// произошла ошибка
			if tt.expectedCode != http.StatusOK && tt.expectedErr != nil {
				var result map[string]string
				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedErr, result)
			} else if tt.expectedCode == http.StatusOK { // успешный кейс
				var result []model.Note

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedResponse, result)
			} else { // у пользователя нет заметок
				assert.Equal(t, tt.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestNotesBySpaceID(t *testing.T) {
	type test struct {
		name             string
		spaceID          string
		dbErr            error // ошибка, которую возвращает база
		expectedCode     int
		expectedResponse []model.GetNote
		expectedErr      map[string]string
	}

	logrus.SetLevel(logrus.DebugLevel)

	wayback := time.Date(2024, time.May, 19, 1, 2, 3, 4, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer patch.Unpatch()

	noteID := uuid.New()
	noteIDPatch := monkey.Patch(uuid.New, func() uuid.UUID { return noteID })
	defer noteIDPatch.Unpatch()

	tests := []test{
		{
			name:         "positive test",
			spaceID:      "1234",
			expectedCode: http.StatusOK,
			expectedResponse: []model.GetNote{
				{
					ID:      uuid.New(),
					UserID:  1,
					Text:    "test note",
					SpaceID: uuid.New(),
					Created: time.Now(),
				},
			},
		},
		{
			name:         "space does not have any notes",
			spaceID:      "1234",
			dbErr:        api_errors.ErrNoNotesFoundBySpaceID,
			expectedCode: http.StatusNoContent,
			expectedErr:  nil,
		},
		{
			name:         "space does not exist",
			spaceID:      "1234",
			dbErr:        api_errors.ErrSpaceNotExists,
			expectedCode: http.StatusNotFound,
			expectedErr:  nil,
		},
		{
			name:         "invalid param",
			spaceID:      "1234abc",
			expectedCode: http.StatusBadRequest,
			expectedErr:  map[string]string{"bad request": "invalid space id parameter: strconv.Atoi: parsing \"1234abc\": invalid syntax"},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockspaceRepo(ctrl)

	spaceSrv := space.New(repo, nil, nil)

	server := New("", spaceSrv, nil)

	r, err := runTestServer(server)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbErr != nil {
				repo.EXPECT().GetAllNotesBySpaceID(gomock.Any(), gomock.Any()).Return(nil, tt.dbErr)
			} else if tt.expectedCode == http.StatusOK {
				repo.EXPECT().GetAllNotesBySpaceID(gomock.Any(), gomock.Any()).Return(tt.expectedResponse, nil)
			}

			url := fmt.Sprintf("/spaces/%s/notes", tt.spaceID)

			resp := testRequest(t, ts, http.MethodGet, url, nil)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			// произошла ошибка
			if tt.expectedCode != http.StatusOK && tt.expectedErr != nil {
				var result map[string]string
				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedErr, result)
			} else if tt.expectedCode == http.StatusOK { // успешный кейс
				var result []model.GetNote

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedResponse, result)
			} else { // у пользователя нет заметок
				assert.Equal(t, tt.expectedCode, resp.StatusCode)
			}
		})
	}
}
