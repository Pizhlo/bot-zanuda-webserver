package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"webserver/internal/model"
	"webserver/internal/service/note"
	"webserver/internal/service/space"
	note_db "webserver/internal/service/storage/postgres/note"
	space_db "webserver/internal/service/storage/postgres/space"
	"webserver/mocks"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateNote(t *testing.T) {
	type test struct {
		name             string
		req              model.CreateNoteRequest
		dbErr            error // ошибка, которую возвращает база
		expectedCode     int
		expectedResponse map[string]string
	}

	tests := []test{
		{
			name:  "positive test",
			dbErr: nil,
			req: model.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: 1,
				Created: time.Now().Unix(),
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:  "field `user_id` not filled",
			dbErr: nil,
			req: model.CreateNoteRequest{
				Text:    "new note",
				SpaceID: 1,
				Created: time.Now().Unix(),
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: map[string]string{"bad request": "field `user_id` not filled"},
		},
		{
			name:  "field `text` not filled",
			dbErr: nil,
			req: model.CreateNoteRequest{
				UserID:  1,
				SpaceID: 1,
				Created: time.Now().Unix(),
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: map[string]string{"bad request": "field `text` not filled"},
		},
		{
			name:  "field `created` not filled",
			dbErr: nil,
			req: model.CreateNoteRequest{
				UserID:  1,
				SpaceID: 1,
				Text:    "new note",
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: map[string]string{"bad request": "field `created` not filled"},
		},
		{
			name:  "field `space_id` not filled",
			dbErr: nil,
			req: model.CreateNoteRequest{
				UserID:  1,
				Created: time.Now().Unix(),
				Text:    "new note",
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: map[string]string{"bad request": "field `space_id` not filled"},
		},
		{
			name:  "db err: unknown user",
			dbErr: note_db.ErrUnknownUser,
			req: model.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: 1,
				Created: time.Now().Unix(),
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: map[string]string{"bad request": "unknown user"},
		},
		{
			name:  "db err: space not exists",
			dbErr: note_db.ErrSpaceNotExists,
			req: model.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: 1,
				Created: time.Now().Unix(),
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: map[string]string{"bad request": "space not exists"},
		},
		{
			name:  "db err: space not exists",
			dbErr: errors.New("unknown error"),
			req: model.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: 1,
				Created: time.Now().Unix(),
			},
			expectedCode:     http.StatusInternalServerError,
			expectedResponse: map[string]string{"error": "unknown error"},
		},
		{
			name:  "db err: space belongs another user",
			dbErr: note_db.ErrSpaceNotBelongsUser,
			req: model.CreateNoteRequest{
				UserID:  1,
				Text:    "new note",
				SpaceID: 1,
				Created: time.Now().Unix(),
			},
			expectedCode:     http.StatusBadRequest,
			expectedResponse: map[string]string{"bad request": "space not belongs to user"},
		},
	}

	ctrl := gomock.NewController(t)
	noteRepo := mocks.NewMocknoteRepo(ctrl)

	noteSrv := note.New(noteRepo)

	server := New("", noteSrv, nil)

	r, err := runTestServer(server)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ожидаем вызова базы либо, если статус кода успешен (значит запрос успешно прошел),
			// либо если есть ошибка из базы
			if tt.dbErr != nil || tt.expectedCode == http.StatusCreated {
				noteRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(tt.dbErr)
			}

			bodyJSON, err := json.Marshal(tt.req)
			require.NoError(t, err)

			resp := testRequest(t, ts, http.MethodPost, "/spaces/notes/create", bytes.NewReader(bodyJSON))
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.expectedCode != http.StatusCreated {
				var result map[string]string
				dec := json.NewDecoder(resp.Body)
				err = dec.Decode(&result)
				require.NoError(t, err)

				assert.Equal(t, tt.expectedResponse, result)
			}
		})
	}
}

func TestNotesBeSpaceID_Full(t *testing.T) {
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
							ID:       1,
							Name:     "personal space for user 1234",
							Created:  time.Now(),
							Creator:  1,
							Personal: true,
						},
						Timezone: "Europe/Moscow",
					},
					Text: "test note",
					Space: &model.Space{
						ID:       1,
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
			dbErr:        space_db.ErrNoNotesFoundBySpaceID,
			expectedCode: http.StatusNoContent,
			expectedErr:  nil,
		},
		{
			name:         "space does not exist",
			spaceID:      "1234",
			dbErr:        space_db.ErrSpaceNotExists,
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
	repo := mocks.NewMockspaceRepo(ctrl)

	spaceSrv := space.New(repo)

	server := New("", nil, spaceSrv)

	r, err := runTestServer(server)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbErr != nil {
				repo.EXPECT().GetAllbySpaceIDFull(gomock.Any(), gomock.Any()).Return(nil, tt.dbErr)
			} else if tt.expectedCode == http.StatusOK {
				repo.EXPECT().GetAllbySpaceIDFull(gomock.Any(), gomock.Any()).Return(tt.expectedResponse, nil)
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

func TestNotesBeSpaceID(t *testing.T) {
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
			spaceID:      "1234",
			expectedCode: http.StatusOK,
			expectedResponse: []model.GetNote{
				{
					ID:       uuid.New(),
					UserID:   1,
					Text:     "test note",
					SpaceID:  1,
					Created:  time.Now(),
					LastEdit: sql.NullTime{Valid: false},
				},
			},
		},
		{
			name:         "space does not have any notes",
			spaceID:      "1234",
			dbErr:        space_db.ErrNoNotesFoundBySpaceID,
			expectedCode: http.StatusNoContent,
			expectedErr:  nil,
		},
		{
			name:         "space does not exist",
			spaceID:      "1234",
			dbErr:        space_db.ErrSpaceNotExists,
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
	repo := mocks.NewMockspaceRepo(ctrl)

	spaceSrv := space.New(repo)

	server := New("", nil, spaceSrv)

	r, err := runTestServer(server)
	require.NoError(t, err)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbErr != nil {
				repo.EXPECT().GetAllBySpaceID(gomock.Any(), gomock.Any()).Return(nil, tt.dbErr)
			} else if tt.expectedCode == http.StatusOK {
				repo.EXPECT().GetAllBySpaceID(gomock.Any(), gomock.Any()).Return(tt.expectedResponse, nil)
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
