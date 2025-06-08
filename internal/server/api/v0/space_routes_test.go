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

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateSpace(t *testing.T) {
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
			name: "positive case",
			req: rabbit.CreateSpaceRequest{
				Name: "test space",
			},
			tokenArgs: tokenArgs{
				userID:  userID,
				expired: float64(time.Now().Add(time.Hour * 24).Unix()),
			},
			setupMocks: func(mocks *fields) {
				t.Helper()

				mocks.authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				mocks.authSrv.EXPECT().GetPayload(gomock.Any()).Return(map[string]interface{}{
					"user_id": userID,
					"expired": float64(1780428974),
				}, true)
				mocks.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
				mocks.spaceSrv.EXPECT().CreateSpace(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "empty space name",
			req: rabbit.CreateSpaceRequest{
				Name: "",
			},
			tokenArgs: tokenArgs{
				userID:  userID,
				expired: float64(time.Now().Add(time.Hour * 24).Unix()),
			},
			setupMocks: func(mocks *fields) {
				t.Helper()
				mocks.authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				mocks.authSrv.EXPECT().GetPayload(gomock.Any()).Return(map[string]interface{}{
					"user_id": userID,
					"expired": float64(1780428974),
				}, true)
				mocks.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
				mocks.spaceSrv.EXPECT().CreateSpace(gomock.Any(), gomock.Any()).Return(model.ErrFieldNameNotFilled)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  model.ErrFieldNameNotFilled,
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

func TestAddParticipant(t *testing.T) {
	fromUser := float64(123)
	toUser := float64(456)
	spaceID := uuid.New()

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
		req            rabbit.AddParticipantRequest
		tokenArgs      tokenArgs
		setupMocks     func(mocks *fields)
		expectedStatus int
		expectedError  *api_errors.HTTPError
	}

	tests := []test{
		{
			name: "positive case",
			req: rabbit.AddParticipantRequest{
				Participant: int64(toUser),
			},
			tokenArgs: tokenArgs{
				userID:  fromUser,
				expired: float64(time.Now().Add(time.Hour * 24).Unix()),
			},
			setupMocks: func(mocks *fields) {
				t.Helper()

				mocks.authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				mocks.authSrv.EXPECT().GetPayload(gomock.Any()).Return(map[string]interface{}{
					"user_id": fromUser,
					"expired": float64(1780428974),
				}, true)
				mocks.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil).Times(2)
				mocks.spaceSrv.EXPECT().IsSpacePersonal(gomock.Any(), spaceID).Return(false, nil)
				mocks.spaceSrv.EXPECT().IsUserInSpace(gomock.Any(), gomock.Any(), spaceID).Return(true, nil)
				mocks.spaceSrv.EXPECT().IsUserInSpace(gomock.Any(), gomock.Any(), spaceID).Return(false, nil)
				mocks.spaceSrv.EXPECT().CheckInvitation(gomock.Any(), gomock.Any(), gomock.Any(), spaceID).Return(false, nil)
				mocks.spaceSrv.EXPECT().AddParticipant(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "trying to add self as participant",
			req: rabbit.AddParticipantRequest{
				Participant: int64(fromUser),
			},
			tokenArgs: tokenArgs{
				userID:  fromUser,
				expired: float64(time.Now().Add(time.Hour * 24).Unix()),
			},
			setupMocks: func(mocks *fields) {
				t.Helper()

				mocks.authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				mocks.authSrv.EXPECT().GetPayload(gomock.Any()).Return(map[string]interface{}{
					"user_id": fromUser,
					"expired": float64(1780428974),
				}, true)
				mocks.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  api_errors.NewHTTPError(http.StatusBadRequest, "you can't add yourself as a participant", nil),
		},
		{
			name: "personal space",
			req: rabbit.AddParticipantRequest{
				Participant: int64(toUser),
			},
			tokenArgs: tokenArgs{
				userID:  fromUser,
				expired: float64(time.Now().Add(time.Hour * 24).Unix()),
			},
			setupMocks: func(mocks *fields) {
				t.Helper()

				mocks.authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				mocks.authSrv.EXPECT().GetPayload(gomock.Any()).Return(map[string]interface{}{
					"user_id": fromUser,
					"expired": float64(1780428974),
				}, true)
				mocks.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil).Times(2)
				mocks.spaceSrv.EXPECT().IsSpacePersonal(gomock.Any(), spaceID).Return(true, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  api_errors.NewHTTPError(http.StatusBadRequest, "personal space", nil),
		},
		{
			name: "space not found",
			req: rabbit.AddParticipantRequest{
				Participant: int64(toUser),
			},
			tokenArgs: tokenArgs{
				userID:  fromUser,
				expired: float64(time.Now().Add(time.Hour * 24).Unix()),
			},
			setupMocks: func(mocks *fields) {
				t.Helper()

				mocks.authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				mocks.authSrv.EXPECT().GetPayload(gomock.Any()).Return(map[string]interface{}{
					"user_id": fromUser,
					"expired": float64(1780428974),
				}, true)
				mocks.userSrv.EXPECT().CheckUser(gomock.Any(), gomock.Any()).Return(true, nil).Times(2)
				mocks.spaceSrv.EXPECT().IsSpacePersonal(gomock.Any(), spaceID).Return(false, api_errors.ErrSpaceNotExists)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  api_errors.NewHTTPError(http.StatusBadRequest, "space not found", nil),
		},
	}

	urlFmt := "/api/v0/spaces/%s/participants/add"

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

			resp := testRequest(t, ts, http.MethodPost, fmt.Sprintf(urlFmt, spaceID), token, bytes.NewBuffer(reqBody))
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
