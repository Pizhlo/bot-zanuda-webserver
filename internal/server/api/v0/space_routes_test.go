package v0

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateSpace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceSrv, userSrv, authSrv := createMockServices(ctrl)

	handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
	require.NoError(t, err)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	validToken := "valid_token"
	userID := float64(123)

	type test struct {
		name           string
		req            rabbit.CreateSpaceRequest
		token          string
		setupMocks     func()
		expectedStatus int
		expectedError  string
	}

	tests := []test{
		{
			name: "positive case",
			req: rabbit.CreateSpaceRequest{
				Name: "test space",
			},
			token: validToken,
			setupMocks: func() {
				authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				authSrv.EXPECT().GetPayload(gomock.Any()).Return(map[string]interface{}{
					"user_id": userID,
				}, true)
				spaceSrv.EXPECT().CreateSpace(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "no auth token",
			req: rabbit.CreateSpaceRequest{
				Name: "test space",
			},
			token:          "",
			setupMocks:     func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "token not found",
		},
		{
			name: "invalid token",
			req: rabbit.CreateSpaceRequest{
				Name: "test space",
			},
			token: "invalid_token",
			setupMocks: func() {
				authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, errors.New("invalid token"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid token",
		},
		{
			name: "no payload in token",
			req: rabbit.CreateSpaceRequest{
				Name: "test space",
			},
			token: validToken,
			setupMocks: func() {
				authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				authSrv.EXPECT().GetPayload(gomock.Any()).Return(nil, false)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "payload in token not found",
		},
		{
			name: "empty space name",
			req: rabbit.CreateSpaceRequest{
				Name: "",
			},
			token: validToken,
			setupMocks: func() {
				authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				authSrv.EXPECT().GetPayload(gomock.Any()).Return(map[string]interface{}{
					"user_id": userID,
				}, true)
				spaceSrv.EXPECT().CreateSpace(gomock.Any(), gomock.Any()).Return(model.ErrFieldNameNotFilled)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  model.ErrFieldNameNotFilled.Error(),
		},
		{
			name: "no user ID in payload",
			req: rabbit.CreateSpaceRequest{
				Name: "test space",
			},
			token: validToken,
			setupMocks: func() {
				authSrv.EXPECT().CheckToken(gomock.Any()).Return(&jwt.Token{}, nil)
				authSrv.EXPECT().GetPayload(gomock.Any()).Return(map[string]interface{}{}, true)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "user id not found in payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			reqBody, err := json.Marshal(tt.req)
			require.NoError(t, err)

			resp := testRequest(t, ts, http.MethodPost, "/api/v0/spaces/create", tt.token, bytes.NewBuffer(reqBody))
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != "" {
				var respBody map[string]string
				err = json.NewDecoder(resp.Body).Decode(&respBody)
				require.NoError(t, err)

				var errorMsg string
				for _, v := range respBody {
					errorMsg = v
					break
				}
				require.Equal(t, tt.expectedError, errorMsg)
			} else {
				var respBody map[string]string
				err = json.NewDecoder(resp.Body).Decode(&respBody)
				require.NoError(t, err)

				_, err := uuid.Parse(respBody["req_id"])
				require.NoError(t, err)
			}
		})
	}
}
