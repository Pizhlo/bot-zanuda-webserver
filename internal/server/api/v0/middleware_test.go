package v0

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"
	"webserver/internal/server/api/v0/mocks"

	"bou.ke/monkey"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/undefinedlabs/go-mpatch"
)

// тест для проверки middleware
func TestValidateNoteRequest_CreateNote(t *testing.T) {
	type fields struct {
		spaceSrv *mocks.MockspaceService
		userSrv  *mocks.MockuserService
		authSrv  *mocks.MockauthService
	}

	type test struct {
		name         string
		req          rabbit.CreateNoteRequest
		expectedNote rabbit.CreateNoteRequest
		// ошибки разных репозиториев
		err              error // ошибки валидации и т.п.
		expectedCode     int
		expectedResponse error
		setupMocks       func(mocks *fields)
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
			setupMocks: func(m *fields) {
				t.Helper()

				m.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsSpaceExists(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsUserInSpace(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(nil).Do(func(ctx any, actualReq rabbit.CreateNoteRequest) {
					expectedNote := rabbit.CreateNoteRequest{
						ID:        uuid.New(),
						UserID:    1,
						Text:      "new note",
						SpaceID:   uuid.New(),
						Type:      model.TextNoteType,
						Created:   time.Now().Unix(),
						Operation: rabbit.CreateOp,
					}
					assert.Equal(t, expectedNote, actualReq, "requests not equal")
				})
			},
		},
		{
			name: "db err: unknown user",
			err:  api_errors.ErrUnknownUser,
			req: rabbit.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: api_errors.ErrUnknownUser,
			setupMocks: func(m *fields) {
				t.Helper()

				m.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(false, nil)
			},
		},
		{
			name: "db err: space not exists",
			err:  api_errors.ErrSpaceNotExists,
			req: rabbit.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: api_errors.ErrSpaceNotExists,
			setupMocks: func(m *fields) {
				t.Helper()

				m.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsSpaceExists(gomock.Any(), gomock.Any()).Return(false, nil)
			},
		},
		{
			name: "db err: space belongs another user",
			err:  api_errors.ErrSpaceNotBelongsUser,
			req: rabbit.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: api_errors.ErrSpaceNotBelongsUser,
			setupMocks: func(m *fields) {
				t.Helper()

				m.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsSpaceExists(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsUserInSpace(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			spaceSrvMock, userSrvMock, authSrvMock := createMockServices(t, ctrl)

			handler, err := New(WithSpaceService(spaceSrvMock), WithUserService(userSrvMock), WithAuthService(authSrvMock))
			require.NoError(t, err)

			r, err := runTestServerWithMiddleware(t, handler)
			require.NoError(t, err)

			ts := httptest.NewServer(r)
			defer ts.Close()

			tt.setupMocks(&fields{
				spaceSrv: spaceSrvMock,
				userSrv:  userSrvMock,
				authSrv:  authSrvMock,
			})

			bodyJSON, err := json.Marshal(tt.req)
			require.NoError(t, err)

			resp := testRequest(t, ts, http.MethodPost, "/api/v0/spaces/notes/create", "", bytes.NewReader(bodyJSON))
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedResponse != nil {
				checkResult(t, resp, tt.expectedResponse)
			} else {
				checkRequestID(t, resp)
			}
		})
	}
}

func TestValidateNoteRequest_UpdateNote(t *testing.T) {
	type fields struct {
		spaceSrv *mocks.MockspaceService
		userSrv  *mocks.MockuserService
		authSrv  *mocks.MockauthService
	}

	type test struct {
		name         string
		req          rabbit.UpdateNoteRequest
		expectedNote rabbit.UpdateNoteRequest
		setupMocks   func(mocks *fields)
		expectedCode int
		expectedErr  error
	}

	generatedID := uuid.New()

	uuidPatch, err := mpatch.PatchMethod(uuid.New, func() uuid.UUID { return generatedID })
	require.NoError(t, err)

	defer uuidPatch.Unpatch()

	wayback := time.Now()
	timePatch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer timePatch.Unpatch()

	note := model.GetNote{
		UserID:  1,
		Text:    "new note",
		SpaceID: generatedID,
		ID:      generatedID,
		Type:    model.TextNoteType,
	}

	tests := []test{
		{
			name: "update note",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: generatedID,
				NoteID:  generatedID,
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
			setupMocks: func(m *fields) {
				t.Helper()

				m.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsSpaceExists(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsUserInSpace(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().GetNoteByID(gomock.Any(), gomock.Any()).Return(note, nil)
				m.spaceSrv.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(nil).Do(func(ctx any, actualReq rabbit.UpdateNoteRequest) {
					expectedNote := rabbit.UpdateNoteRequest{
						ID:        generatedID,
						UserID:    1,
						Text:      "new note",
						SpaceID:   generatedID,
						NoteID:    generatedID,
						Created:   time.Now().Unix(),
						Operation: rabbit.UpdateOp,
					}
					assert.Equal(t, expectedNote, actualReq, "requests not equal")
				})
			},
		},
		{
			name: "db err: unknown user",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  api_errors.ErrUnknownUser,
			setupMocks: func(m *fields) {
				t.Helper()

				m.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(false, nil)
			},
		},
		{
			name: "db err: space not exists",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  api_errors.ErrSpaceNotExists,
			setupMocks: func(m *fields) {
				t.Helper()

				m.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsSpaceExists(gomock.Any(), gomock.Any()).Return(false, nil)
			},
		},
		{
			name: "db err: space belongs another user",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  api_errors.ErrSpaceNotBelongsUser,
			setupMocks: func(m *fields) {
				t.Helper()

				m.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsSpaceExists(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsUserInSpace(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			spaceSrv, userSrv, authSrv := createMockServices(t, ctrl)

			handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
			require.NoError(t, err)

			r, err := runTestServerWithMiddleware(t, handler)
			require.NoError(t, err)

			ts := httptest.NewServer(r)
			defer ts.Close()

			tt.setupMocks(&fields{
				spaceSrv: spaceSrv,
				userSrv:  userSrv,
				authSrv:  authSrv,
			})

			bodyJSON, err := json.Marshal(tt.req)
			require.NoError(t, err)

			resp := testRequest(t, ts, http.MethodPatch, "/api/v0/spaces/notes/update", "", bytes.NewReader(bodyJSON))
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedErr != nil {
				checkResult(t, resp, tt.expectedErr)
			} else {
				checkRequestID(t, resp)
			}
		})
	}
}

func TestWrapNetHTTP(t *testing.T) {
	type fields struct {
		spaceSrv *mocks.MockspaceService
		userSrv  *mocks.MockuserService
		authSrv  *mocks.MockauthService
	}

	type test struct {
		name         string
		req          rabbit.UpdateNoteRequest
		expectedNote rabbit.UpdateNoteRequest
		setupMocks   func(mocks *fields)
		expectedCode int
		expectedErr  error
	}

	generatedID := uuid.New()

	uuidPatch, err := mpatch.PatchMethod(uuid.New, func() uuid.UUID { return generatedID })
	require.NoError(t, err)

	defer uuidPatch.Unpatch()

	wayback := time.Now()
	timePatch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer timePatch.Unpatch()

	note := model.GetNote{
		UserID:  1,
		Text:    "new note",
		SpaceID: generatedID,
		ID:      generatedID,
		Type:    model.TextNoteType,
	}

	tests := []test{
		{
			name: "update note",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: generatedID,
				NoteID:  generatedID,
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
			setupMocks: func(m *fields) {
				t.Helper()

				m.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsSpaceExists(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsUserInSpace(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().GetNoteByID(gomock.Any(), gomock.Any()).Return(note, nil)
				m.spaceSrv.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(nil).Do(func(ctx any, actualReq rabbit.UpdateNoteRequest) {
					expectedNote := rabbit.UpdateNoteRequest{
						ID:        generatedID,
						UserID:    1,
						Text:      "new note",
						SpaceID:   generatedID,
						NoteID:    generatedID,
						Created:   time.Now().Unix(),
						Operation: rabbit.UpdateOp,
					}
					assert.Equal(t, expectedNote, actualReq, "requests not equal")
				})
			},
		},
		{
			name: "db err: unknown user",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  api_errors.ErrUnknownUser,
			setupMocks: func(m *fields) {
				t.Helper()

				m.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(false, nil)
			},
		},
		{
			name: "db err: space not exists",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  api_errors.ErrSpaceNotExists,
			setupMocks: func(m *fields) {
				t.Helper()

				m.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsSpaceExists(gomock.Any(), gomock.Any()).Return(false, nil)
			},
		},
		{
			name: "db err: space belongs another user",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  api_errors.ErrSpaceNotBelongsUser,
			setupMocks: func(m *fields) {
				t.Helper()

				m.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsSpaceExists(gomock.Any(), gomock.Any()).Return(true, nil)
				m.spaceSrv.EXPECT().IsUserInSpace(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			spaceSrv, userSrv, authSrv := createMockServices(t, ctrl)

			handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
			require.NoError(t, err)

			r, err := runTestServerWithMiddleware(t, handler)
			require.NoError(t, err)

			ts := httptest.NewServer(r)
			defer ts.Close()

			tt.setupMocks(&fields{
				spaceSrv: spaceSrv,
				userSrv:  userSrv,
				authSrv:  authSrv,
			})

			bodyJSON, err := json.Marshal(tt.req)
			require.NoError(t, err)

			resp := testRequest(t, ts, http.MethodPatch, "/api/v0/spaces/notes/update", "", bytes.NewReader(bodyJSON))
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedErr != nil {
				checkResult(t, resp, tt.expectedErr)
			} else {
				checkRequestID(t, resp)
			}
		})
	}
}

func TestAuth(t *testing.T) {
	type fields struct {
		spaceSrv *mocks.MockspaceService
		userSrv  *mocks.MockuserService
		authSrv  *mocks.MockauthService
	}

	type tokenArgs struct {
		userID  float64
		expired float64
	}

	type test struct {
		name           string
		req            rabbit.CreateSpaceRequest
		tokenArgs      tokenArgs
		setupMocks     func(mocks *fields)
		expectedStatus int
		expectedError  error
	}

	userID := float64(123)

	tests := []test{
		{
			name: "no payload in token",
			req: rabbit.CreateSpaceRequest{
				Name: "test space",
			},
			setupMocks: func(m *fields) {
				t.Helper()

				m.authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				m.authSrv.EXPECT().GetPayload(gomock.Any()).Return(jwt.MapClaims{}, false)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  api_errors.ErrNoPayloadInToken,
		},
		{
			name: "user not found",
			req: rabbit.CreateSpaceRequest{
				Name: "test space",
			},
			tokenArgs: tokenArgs{
				expired: float64(time.Now().Add(time.Hour * 24).Unix()),
			},
			setupMocks: func(m *fields) {
				t.Helper()

				m.authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				m.authSrv.EXPECT().GetPayload(gomock.Any()).Return(jwt.MapClaims{
					"expired": float64(time.Now().Add(time.Hour * 24).Unix()),
				}, true)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  api_errors.ErrUserNotFoundInPayload,
		},
		{
			name: "expired not found in payload",
			req: rabbit.CreateSpaceRequest{
				Name: "test space",
			},
			tokenArgs: tokenArgs{
				userID: userID,
			},
			setupMocks: func(m *fields) {
				t.Helper()

				m.authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				m.authSrv.EXPECT().GetPayload(gomock.Any()).Return(jwt.MapClaims{
					"user_id": userID,
				}, true)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  api_errors.ErrExpiredNotFoundInPayload,
		},
		{
			name: "token expired",
			req: rabbit.CreateSpaceRequest{
				Name: "test space",
			},
			tokenArgs: tokenArgs{
				userID:  userID,
				expired: float64(time.Now().Add(-time.Hour * 24).Unix()),
			},
			setupMocks: func(m *fields) {
				t.Helper()

				m.authSrv.EXPECT().CheckToken(gomock.Any()).Return(nil, api_errors.ErrTokenExpired)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  api_errors.ErrTokenExpired,
		},
		{
			name: "user not exists",
			req: rabbit.CreateSpaceRequest{
				Name: "test space",
			},
			tokenArgs: tokenArgs{
				userID:  userID,
				expired: float64(time.Now().Add(time.Hour * 24).Unix()),
			},
			setupMocks: func(m *fields) {
				t.Helper()

				m.authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				m.authSrv.EXPECT().GetPayload(gomock.Any()).Return(jwt.MapClaims{
					"user_id": userID,
					"expired": float64(time.Now().Add(time.Hour * 24).Unix()),
				}, true)
				m.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(false, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  fmt.Errorf("user %d not found", int64(userID)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			spaceSrv, userSrv, authSrv := createMockServices(t, ctrl)

			handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
			require.NoError(t, err)

			r, err := runTestServer(t, handler)
			require.NoError(t, err)

			ts := httptest.NewServer(r)
			defer ts.Close()

			tt.setupMocks(&fields{spaceSrv: spaceSrv, userSrv: userSrv, authSrv: authSrv})

			reqBody, err := json.Marshal(tt.req)
			require.NoError(t, err)

			token := generateToken(t, tt.tokenArgs.userID, tt.tokenArgs.expired)

			resp := testRequest(t, ts, http.MethodPost, "/api/v0/spaces/create", token, bytes.NewBuffer(reqBody))
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != nil {
				checkResult(t, resp, tt.expectedError)
			} else {
				checkRequestID(t, resp)
			}
		})
	}
}
