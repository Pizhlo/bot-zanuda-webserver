package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"webserver/internal/model"
	"webserver/internal/service/note"
	note_db "webserver/internal/service/storage/postgres/note"
	"webserver/mocks"

	"github.com/golang/mock/gomock"
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
	repo := mocks.NewMockrepo(ctrl)

	noteSrv := note.New(repo)

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
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(tt.dbErr)
			}

			bodyJSON, err := json.Marshal(tt.req)
			require.NoError(t, err)

			resp := testRequest(t, ts, http.MethodPost, "/notes/create", bytes.NewReader(bodyJSON))
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
