package v0

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
	"webserver/internal/model/rabbit"

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
		req              rabbit.CreateNoteRequest
		expectedNote     rabbit.CreateNoteRequest
		expectedErr      error
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
			name: "positive test",
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
			name: "user ID not filled",
			req: rabbit.CreateNoteRequest{
				Text:    "new note",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			expectedNote: rabbit.CreateNoteRequest{
				ID:        uuid.New(),
				Text:      "new note",
				SpaceID:   uuid.New(),
				Type:      model.TextNoteType,
				Created:   time.Now().Unix(),
				Operation: rabbit.CreateOp,
			},
			expectedErr:  model.ErrFieldUserNotFilled,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrFieldUserNotFilled.Error(),
			},
		},
		{
			name: "text not filled",
			req: rabbit.CreateNoteRequest{
				UserID:  1,
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
			},
			expectedNote: rabbit.CreateNoteRequest{
				ID:        uuid.New(),
				UserID:    1,
				SpaceID:   uuid.New(),
				Type:      model.TextNoteType,
				Created:   time.Now().Unix(),
				Operation: rabbit.CreateOp,
			},
			expectedErr:  model.ErrFieldTextNotFilled,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrFieldTextNotFilled.Error(),
			},
		},
		{
			name: "invalid space ID",
			req: rabbit.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.Nil,
				Type:    model.TextNoteType,
			},
			expectedNote: rabbit.CreateNoteRequest{
				ID:        uuid.New(),
				UserID:    1,
				Text:      "new note",
				SpaceID:   uuid.Nil,
				Type:      model.TextNoteType,
				Created:   time.Now().Unix(),
				Operation: rabbit.CreateOp,
			},
			expectedErr:  model.ErrInvalidSpaceID,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrInvalidSpaceID.Error(),
			},
		},
		{
			name: "field type not filled",
			req: rabbit.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: uuid.New(),
			},
			expectedNote: rabbit.CreateNoteRequest{
				ID:        uuid.New(),
				UserID:    1,
				Text:      "new note",
				SpaceID:   uuid.New(),
				Created:   time.Now().Unix(),
				Operation: rabbit.CreateOp,
			},
			expectedErr:  model.ErrFieldTypeNotFilled,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrFieldTypeNotFilled.Error(),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceSrv, userSrv, authSrv := createMockServices(ctrl)

	handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
	require.NoError(t, err)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, err := json.Marshal(tt.req)
			require.NoError(t, err)

			if tt.expectedErr == nil {
				spaceSrv.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(nil)
			} else {
				spaceSrv.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(tt.expectedErr)
			}

			resp := testRequest(t, ts, http.MethodPost, "/api/v0/spaces/notes/create", bytes.NewReader(bodyJSON))
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedResponse != nil {
				var result map[string]string
				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedResponse, result)
			}
		})
	}
}

func TestUpdateNote(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	type test struct {
		name             string
		req              rabbit.UpdateNoteRequest
		expectedNote     rabbit.UpdateNoteRequest
		dbNote           model.GetNote // что возвращает база при вызове GetNote
		dbErr            bool
		err              error
		getNote          bool // нужно ли вызывать getNote
		expectedCode     int
		expectedResponse map[string]string
	}

	generatedID := uuid.New()
	newID := uuid.New() // для случая, когда айди должен отличаться (заметка не принадлежит пространству)

	wayback := time.Now()
	timePatch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer timePatch.Unpatch()

	tests := []test{
		{
			name: "positive test",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: generatedID,
				NoteID:  generatedID,
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
			expectedNote: rabbit.UpdateNoteRequest{
				UserID:    1,
				ID:        generatedID,
				Text:      "new note",
				SpaceID:   generatedID,
				NoteID:    generatedID,
				Created:   time.Now().Unix(),
				Operation: rabbit.UpdateOp,
			},
			expectedCode: http.StatusAccepted,
		},
		{
			name: "user ID not filled",
			req: rabbit.UpdateNoteRequest{
				Text:    "new note",
				SpaceID: uuid.New(),
			},
			dbErr:        true,
			err:          model.ErrFieldUserNotFilled,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrFieldUserNotFilled.Error(),
			},
		},
		{
			name: "text not filled",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				SpaceID: uuid.New(),
			},
			err:          model.ErrFieldTextNotFilled,
			dbErr:        true,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrFieldTextNotFilled.Error(),
			},
		},
		{
			name: "invalid space ID",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "invalid space ID",
				SpaceID: uuid.Nil,
				NoteID:  generatedID,
			},
			err:          model.ErrInvalidSpaceID,
			dbErr:        true,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": model.ErrInvalidSpaceID.Error(),
			},
		},
		{
			name: "note not belongs space",
			req: rabbit.UpdateNoteRequest{
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
			err:          api_errors.ErrNoteNotBelongsSpace,
			getNote:      true,
			dbErr:        false,
			expectedCode: http.StatusBadRequest,
			expectedResponse: map[string]string{
				"bad request": api_errors.ErrNoteNotBelongsSpace.Error(),
			},
		},
		{
			name: "note type is not text",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "note type is not text REQ",
				SpaceID: generatedID,
				NoteID:  generatedID,
			},
			err:     model.ErrUpdateNotTextNote,
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
		{
			name: "note not found",
			req: rabbit.UpdateNoteRequest{
				UserID:  1,
				Text:    "note not found REQ",
				SpaceID: generatedID,
				NoteID:  generatedID,
			},
			dbErr:        true,
			err:          api_errors.ErrNoteNotFound,
			expectedCode: http.StatusNotFound,
			expectedResponse: map[string]string{
				"error": api_errors.ErrNoteNotFound.Error(),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceSrv, userSrv, authSrv := createMockServices(ctrl)

	handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
	require.NoError(t, err)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbErr {
				spaceSrv.EXPECT().GetNoteByID(gomock.Any(), gomock.Any()).Return(tt.dbNote, tt.err)
			} else {
				spaceSrv.EXPECT().GetNoteByID(gomock.Any(), gomock.Any()).Return(tt.dbNote, nil)
			}

			if tt.err == nil {
				spaceSrv.EXPECT().UpdateNote(gomock.Any(), gomock.Any()).Return(nil)
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

				assert.Equal(t, tt.expectedResponse, result, "responses not equal")
			} else {
				var result map[string]string
				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				// проверяем, что возвращается валидный uuid
				_, err := uuid.Parse(result["request_id"])
				require.NoError(t, err)
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
			spaceID:      uuid.New().String(),
			expectedCode: http.StatusOK,
			expectedResponse: []model.Note{
				{
					ID: uuid.New(),
					User: &model.User{
						ID:       1,
						TgID:     1234,
						Username: "test user",
						PersonalSpace: &model.Space{
							ID:       uuid.New(),
							Name:     "personal space for user 1234",
							Created:  time.Now().Unix(),
							Creator:  1,
							Personal: true,
						},
						Timezone: "Europe/Moscow",
					},
					Text: "test note",
					Space: &model.Space{
						ID:       uuid.New(),
						Name:     "personal space for user 1234",
						Created:  time.Now().Unix(),
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
			spaceID:      uuid.New().String(),
			dbErr:        api_errors.ErrNoNotesFoundBySpaceID,
			expectedCode: http.StatusNotFound,
			expectedErr:  nil,
		},
		{
			name:         "space does not exist",
			spaceID:      uuid.New().String(),
			dbErr:        api_errors.ErrSpaceNotExists,
			expectedCode: http.StatusNotFound,
			expectedErr:  nil,
		},
		{
			name:         "invalid param",
			spaceID:      "1234abc",
			expectedCode: http.StatusBadRequest,
			expectedErr:  map[string]string{"bad request": "invalid space id parameter: invalid UUID length: 7"},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceSrv, userSrv, authSrv := createMockServices(ctrl)

	handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
	require.NoError(t, err)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbErr != nil {
				spaceSrv.EXPECT().GetAllNotesBySpaceIDFull(gomock.Any(), gomock.Any()).Return(nil, tt.dbErr)
			} else if tt.expectedCode == http.StatusOK {
				spaceSrv.EXPECT().GetAllNotesBySpaceIDFull(gomock.Any(), gomock.Any()).Return(tt.expectedResponse, nil)
			}

			url := fmt.Sprintf("/api/v0/spaces/%s/notes?full_user=true", tt.spaceID)

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

	wayback := time.Date(2024, time.May, 19, 1, 2, 3, 4, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer patch.Unpatch()

	noteID := uuid.New()
	noteIDPatch := monkey.Patch(uuid.New, func() uuid.UUID { return noteID })
	defer noteIDPatch.Unpatch()

	tests := []test{
		{
			name:         "positive test",
			spaceID:      uuid.New().String(),
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
			spaceID:      uuid.New().String(),
			dbErr:        api_errors.ErrNoNotesFoundBySpaceID,
			expectedCode: http.StatusNotFound,
			expectedErr:  nil,
		},
		{
			name:         "space does not exist",
			spaceID:      uuid.New().String(),
			dbErr:        api_errors.ErrSpaceNotExists,
			expectedCode: http.StatusNotFound,
			expectedErr:  nil,
		},
		{
			name:         "invalid param",
			spaceID:      "1234abc",
			expectedCode: http.StatusBadRequest,
			expectedErr:  map[string]string{"bad request": "invalid space id parameter: invalid UUID length: 7"},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceSrv, userSrv, authSrv := createMockServices(ctrl)

	handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
	require.NoError(t, err)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbErr != nil {
				spaceSrv.EXPECT().GetAllNotesBySpaceID(gomock.Any(), gomock.Any()).Return(nil, tt.dbErr)
			} else if tt.expectedCode == http.StatusOK {
				spaceSrv.EXPECT().GetAllNotesBySpaceID(gomock.Any(), gomock.Any()).Return(tt.expectedResponse, nil)
			}

			url := fmt.Sprintf("/api/v0/spaces/%s/notes", tt.spaceID)

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

func TestGetNoteTypes(t *testing.T) {
	type test struct {
		name             string
		spaceID          string
		err              error
		expectedCode     int
		expectedResponse []model.NoteTypeResponse
		expectedErr      map[string]string
	}

	tests := []test{
		{
			name:         "positive test",
			spaceID:      uuid.NewString(),
			expectedCode: http.StatusOK,
			expectedResponse: []model.NoteTypeResponse{
				{
					Type:  model.TextNoteType,
					Count: 10,
				},
				{
					Type:  model.PhotoNoteType,
					Count: 1,
				},
			},
		},
		{
			name:         "no notes in space",
			spaceID:      uuid.NewString(),
			expectedCode: http.StatusNotFound,
			err:          api_errors.ErrNoNotesFoundBySpaceID,
		},
		{
			name:         "invalid param",
			spaceID:      "1234abc",
			expectedCode: http.StatusBadRequest,
			expectedErr:  map[string]string{"bad request": "invalid space id parameter: invalid UUID length: 7"},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceSrv, userSrv, authSrv := createMockServices(ctrl)

	handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
	require.NoError(t, err)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				spaceSrv.EXPECT().GetNotesTypes(gomock.Any(), gomock.Any()).Return(nil, tt.err)
			} else if tt.expectedCode == http.StatusOK {
				spaceSrv.EXPECT().GetNotesTypes(gomock.Any(), gomock.Any()).Return(tt.expectedResponse, nil)
			}

			url := fmt.Sprintf("/api/v0/spaces/%s/notes/types", tt.spaceID)

			resp := testRequest(t, ts, http.MethodGet, url, nil)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedCode == http.StatusOK { // успешный кейс
				var result []model.NoteTypeResponse

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedResponse, result)
			} else if tt.expectedErr != nil {
				var result map[string]string

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedErr, result)
			} else { // у пользователя нет заметок
				assert.Equal(t, tt.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestGetNotesByType(t *testing.T) {
	type test struct {
		name             string
		spaceID          string
		noteType         string
		dbErr            error // ошибка, которую возвращает база
		expectedCode     int
		expectedResponse []model.GetNote
		expectedErr      map[string]string
	}

	tests := []test{
		{
			name:         "positive test",
			spaceID:      uuid.NewString(),
			noteType:     string(model.TextNoteType),
			expectedCode: http.StatusOK,
			expectedResponse: []model.GetNote{
				{
					ID:      uuid.New(),
					UserID:  1234,
					Text:    "positive test",
					SpaceID: uuid.New(),
					Type:    model.TextNoteType,
				},
			},
		},
		{
			name:         "no notes in space by type",
			spaceID:      uuid.NewString(),
			expectedCode: http.StatusNotFound,
			dbErr:        api_errors.ErrNoNotesFoundByType,
			noteType:     string(model.TextNoteType),
		},
		{
			name:         "invalid type",
			spaceID:      uuid.NewString(),
			expectedCode: http.StatusBadRequest,
			noteType:     "video",
			expectedErr:  map[string]string{"bad request": "invalid note type: video"},
		},
		{
			name:         "invalid param",
			spaceID:      "1234abc",
			noteType:     "text",
			expectedCode: http.StatusBadRequest,
			expectedErr:  map[string]string{"bad request": "invalid space id parameter: invalid UUID length: 7"},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceSrv, userSrv, authSrv := createMockServices(ctrl)

	handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
	require.NoError(t, err)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbErr != nil {
				spaceSrv.EXPECT().GetNotesByType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, tt.dbErr)
			} else if tt.expectedCode == http.StatusOK {
				spaceSrv.EXPECT().GetNotesByType(gomock.Any(), gomock.Any(), gomock.Any()).Return(tt.expectedResponse, nil)
			}

			url := fmt.Sprintf("/api/v0/spaces/%s/notes/%s", tt.spaceID, tt.noteType)

			resp := testRequest(t, ts, http.MethodGet, url, nil)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedCode == http.StatusOK { // успешный кейс
				var result []model.GetNote

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedResponse, result)
			} else if tt.expectedErr != nil {
				var result map[string]string

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedErr, result)
			} else { // у пользователя нет заметок
				assert.Equal(t, tt.expectedCode, resp.StatusCode)
			}

		})
	}
}

func TestSearchNotesByText(t *testing.T) {
	type test struct {
		name             string
		spaceID          string
		req              model.SearchNoteByTextRequest
		dbErr            error // ошибка, которую возвращает база
		expectedCode     int
		expectedResponse []model.GetNote
		expectedErr      map[string]string
	}

	tests := []test{
		{
			name:         "positive test: without parameter",
			spaceID:      uuid.NewString(),
			expectedCode: http.StatusOK,
			req: model.SearchNoteByTextRequest{
				SpaceID: uuid.New(),
				Text:    "positive test",
			},
			expectedResponse: []model.GetNote{
				{
					ID:      uuid.New(),
					UserID:  1234,
					Text:    "positive test",
					SpaceID: uuid.New(),
					Type:    model.TextNoteType,
				},
			},
		},
		{
			name:         "positive test: with parameter",
			spaceID:      uuid.NewString(),
			expectedCode: http.StatusOK,
			req: model.SearchNoteByTextRequest{
				SpaceID: uuid.New(),
				Text:    "positive test",
				Type:    "text",
			},
			expectedResponse: []model.GetNote{
				{
					ID:      uuid.New(),
					UserID:  1234,
					Text:    "positive test",
					SpaceID: uuid.New(),
					Type:    model.TextNoteType,
				},
			},
		},
		{
			name:         "invalid note type",
			spaceID:      uuid.NewString(),
			expectedCode: http.StatusBadRequest,
			req: model.SearchNoteByTextRequest{
				SpaceID: uuid.New(),
				Text:    "positive test",
				Type:    "video",
			},
			expectedErr: map[string]string{"bad request": "invalid note type: video"},
		},
		{
			name:         "notes not found",
			spaceID:      uuid.NewString(),
			expectedCode: http.StatusNotFound,
			dbErr:        api_errors.ErrNoNotesFoundByText,
			req: model.SearchNoteByTextRequest{
				SpaceID: uuid.New(),
				Text:    "positive test",
				Type:    "text",
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceSrv, userSrv, authSrv := createMockServices(ctrl)

	handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
	require.NoError(t, err)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	url := "/api/v0/spaces/notes/search/text"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbErr != nil {
				spaceSrv.EXPECT().SearchNoteByText(gomock.Any(), gomock.Any()).Return(nil, tt.dbErr)
			} else if tt.expectedCode == http.StatusOK {
				spaceSrv.EXPECT().SearchNoteByText(gomock.Any(), gomock.Any()).Return(tt.expectedResponse, nil)
			}

			bodyJSON, err := json.Marshal(tt.req)
			require.NoError(t, err)

			resp := testRequest(t, ts, http.MethodPost, url, bytes.NewReader(bodyJSON))
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedCode == http.StatusOK { // успешный кейс
				var result []model.GetNote

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedResponse, result)
			} else if tt.expectedErr != nil {
				var result map[string]string

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedErr, result)
			} else { // у пользователя нет заметок
				assert.Equal(t, tt.expectedCode, resp.StatusCode)
			}

		})
	}
}

func TestDeleteNote(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	type test struct {
		name            string
		spaceID, noteID string
		dbNote          model.GetNote            // заметка, возвращаемая базой
		expectedReq     rabbit.DeleteNoteRequest // ожидаемое сообщение для воркера
		expectedCode    int
		expectedErr     map[string]string
	}

	// для валидных тестов
	spaceID := uuid.New()
	noteID := uuid.New()

	wayback := time.Now()
	timePatch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer timePatch.Unpatch()

	uuidPatch := monkey.Patch(uuid.New, func() uuid.UUID { return noteID })
	defer uuidPatch.Unpatch()

	tests := []test{
		{
			name:    "positive case",
			spaceID: spaceID.String(),
			noteID:  noteID.String(),
			dbNote: model.GetNote{
				ID:      noteID,
				SpaceID: spaceID,
			},
			expectedReq: rabbit.DeleteNoteRequest{
				ID:        uuid.New(),
				SpaceID:   spaceID,
				NoteID:    noteID,
				Created:   time.Now().Unix(),
				Operation: rabbit.DeleteOp,
			},
			expectedCode: http.StatusAccepted,
		},
		{
			name:         "invalid space ID",
			spaceID:      "abc",
			noteID:       noteID.String(),
			expectedErr:  map[string]string{"bad request": "invalid space id parameter: invalid UUID length: 3"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid note ID",
			spaceID:      spaceID.String(),
			noteID:       "abc",
			expectedErr:  map[string]string{"bad request": "invalid note id parameter: invalid UUID length: 3"},
			expectedCode: http.StatusBadRequest,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceSrv, userSrv, authSrv := createMockServices(ctrl)

	handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
	require.NoError(t, err)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	urlFmt := "/api/v0/spaces/%s/notes/%s/delete"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			url := fmt.Sprintf(urlFmt, tt.spaceID, tt.noteID)

			// happy case
			if tt.expectedErr == nil {
				spaceSrv.EXPECT().GetNoteByID(gomock.Any(), gomock.Any()).Return(tt.dbNote, nil).Do(func(ctx any, actualNoteID uuid.UUID) {
					assert.Equal(t, noteID, actualNoteID)
				})

				spaceSrv.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{ID: spaceID}, nil)

				spaceSrv.EXPECT().DeleteNote(gomock.Any(), gomock.Any()).Return(nil)
			}

			resp := testRequest(t, ts, http.MethodDelete, url, nil)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedCode == http.StatusAccepted {
				var result map[string]string // проверяем, что возвращается валидный uuid

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				idStr, ok := result["req_id"]
				assert.True(t, ok)

				_, err = uuid.Parse(idStr)
				require.NoError(t, err)
			} else {
				var result map[string]string

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedErr, result)
			}
		})
	}
}

func TestDeleteNote_Invalid(t *testing.T) {
	type test struct {
		name            string
		spaceID, noteID string
		dbNote          model.GetNote // заметка, возвращаемая базой
		expectedCode    int
		expectedErr     map[string]string
		methodErrors    map[string]error // название метода : ошибка
	}

	// для валидных тестов
	spaceID := uuid.New()
	noteID := uuid.New()

	wayback := time.Now()
	timePatch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer timePatch.Unpatch()

	uuidPatch := monkey.Patch(uuid.New, func() uuid.UUID { return noteID })
	defer uuidPatch.Unpatch()

	tests := []test{
		{
			name:         "space does not exist",
			spaceID:      spaceID.String(),
			noteID:       noteID.String(),
			expectedErr:  map[string]string{"bad request": api_errors.ErrSpaceNotExists.Error()},
			expectedCode: http.StatusBadRequest,
			methodErrors: map[string]error{"GetSpaceByID": api_errors.ErrSpaceNotExists},
		},
		{
			name:         "note not found",
			spaceID:      spaceID.String(),
			noteID:       noteID.String(),
			expectedErr:  map[string]string{"bad request": api_errors.ErrNoteNotFound.Error()},
			expectedCode: http.StatusNotFound,
			methodErrors: map[string]error{"GetNoteByID": api_errors.ErrNoteNotFound},
		},
		{
			name:    "note does not belong space",
			spaceID: spaceID.String(),
			noteID:  noteID.String(),
			dbNote: model.GetNote{
				SpaceID: uuid.New(),
			},
			expectedErr:  map[string]string{"bad request": api_errors.ErrNoteNotBelongsSpace.Error()},
			expectedCode: http.StatusBadRequest,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceSrv, userSrv, authSrv := createMockServices(ctrl)

	handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
	require.NoError(t, err)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	urlFmt := "/api/v0/spaces/%s/notes/%s/delete"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			url := fmt.Sprintf(urlFmt, tt.spaceID, tt.noteID)

			if err, ok := tt.methodErrors["GetSpaceByID"]; ok {
				spaceSrv.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, err).Do(func(ctx any, actualSpaceID uuid.UUID) {
					assert.Equal(t, spaceID, actualSpaceID)
				})
			} else {
				if err, ok := tt.methodErrors["GetNoteByID"]; ok {
					spaceSrv.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, nil).Do(func(ctx any, actualSpaceID uuid.UUID) {
						assert.Equal(t, spaceID, actualSpaceID)
					})

					spaceSrv.EXPECT().GetNoteByID(gomock.Any(), gomock.Any()).Return(model.GetNote{}, err).Do(func(ctx any, actualNoteID uuid.UUID) {
						assert.Equal(t, noteID, actualNoteID)
					})
				} else {
					spaceSrv.EXPECT().GetNoteByID(gomock.Any(), gomock.Any()).Return(tt.dbNote, nil).Do(func(ctx any, actualNoteID uuid.UUID) {
						assert.Equal(t, noteID, actualNoteID)
					})
				}
			}

			if tt.methodErrors == nil {
				spaceSrv.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{ID: spaceID}, nil)
			}

			// happy case
			if tt.expectedErr == nil {
			}

			resp := testRequest(t, ts, http.MethodDelete, url, nil)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedCode == http.StatusAccepted {
				var result map[string]string // проверяем, что возвращается валидный uuid

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				idStr, ok := result["req_id"]
				assert.True(t, ok)

				_, err = uuid.Parse(idStr)
				require.NoError(t, err)
			} else {
				var result map[string]string

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedErr, result)
			}
		})
	}
}

func TestDeleteAllNotes(t *testing.T) {
	type test struct {
		name         string
		spaceID      string
		expectedReq  rabbit.DeleteAllNotesRequest // ожидаемое сообщение для воркера
		expectedCode int
		expectedErr  map[string]string
		methodErrors map[string]error // название метода : ошибка
	}

	spaceID := uuid.New()

	wayback := time.Now()
	timePatch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer timePatch.Unpatch()

	uuidPatch := monkey.Patch(uuid.New, func() uuid.UUID { return spaceID })
	defer uuidPatch.Unpatch()

	tests := []test{
		{
			name:    "positive case",
			spaceID: spaceID.String(),
			expectedReq: rabbit.DeleteAllNotesRequest{
				ID:        uuid.New(),
				SpaceID:   spaceID,
				Operation: rabbit.DeleteAllOp,
				Created:   wayback.Unix(),
			},
			expectedCode: http.StatusAccepted,
		},
		{
			name:         "space does not exist",
			spaceID:      spaceID.String(),
			expectedErr:  map[string]string{"bad request": api_errors.ErrSpaceNotExists.Error()},
			expectedCode: http.StatusBadRequest,
			methodErrors: map[string]error{"GetSpaceByID": api_errors.ErrSpaceNotExists},
		},
		{
			name:         "invalid space ID",
			spaceID:      "abc",
			expectedErr:  map[string]string{"bad request": "invalid space id parameter: invalid UUID length: 3"},
			expectedCode: http.StatusBadRequest,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	spaceSrv, userSrv, authSrv := createMockServices(ctrl)

	handler, err := New(WithSpaceService(spaceSrv), WithUserService(userSrv), WithAuthService(authSrv))
	require.NoError(t, err)

	r, err := runTestServer(handler)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	urlFmt := "/api/v0/spaces/%s/notes/delete_all"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf(urlFmt, tt.spaceID)

			if err, ok := tt.methodErrors["GetSpaceByID"]; ok {
				spaceSrv.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{}, err).Do(func(ctx any, actualSpaceID uuid.UUID) {
					assert.Equal(t, spaceID, actualSpaceID)
				})
			} else if tt.expectedErr == nil {
				spaceSrv.EXPECT().GetSpaceByID(gomock.Any(), gomock.Any()).Return(model.Space{ID: spaceID}, nil).Do(func(ctx any, actualSpaceID uuid.UUID) {
					assert.Equal(t, spaceID, actualSpaceID)
				})
			}

			if err, ok := tt.methodErrors["DeleteAllNotes"]; ok {
				spaceSrv.EXPECT().DeleteAllNotes(gomock.Any(), gomock.Any()).Return(err)
			}

			if tt.methodErrors == nil && tt.expectedErr == nil {
				spaceSrv.EXPECT().DeleteAllNotes(gomock.Any(), gomock.Any()).Return(nil)
			}

			resp := testRequest(t, ts, http.MethodDelete, url, nil)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedCode == http.StatusAccepted {
				var result map[string]string // проверяем, что возвращается валидный uuid

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				idStr, ok := result["req_id"]
				assert.True(t, ok)

				_, err = uuid.Parse(idStr)
				require.NoError(t, err)
			} else {
				var result map[string]string

				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedErr, result)
			}
		})
	}
}
