package space

import (
	"context"
	"errors"
	"testing"
	"time"
	"webserver/internal/model"
	"webserver/internal/model/rabbit"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateNote(t *testing.T) {
	type test struct {
		name string
		req  rabbit.CreateNoteRequest
		err  error
	}

	tests := []test{
		{
			name: "positive case",
			req: rabbit.CreateNoteRequest{
				ID:        uuid.New(),
				Created:   1236788,
				UserID:    123,
				SpaceID:   uuid.New(),
				Operation: rabbit.CreateOp,
				Text:      "test",
				Type:      model.TextNoteType,
			},
			err: nil,
		},
		{
			name: "error case: db error",
			req: rabbit.CreateNoteRequest{
				ID:        uuid.New(),
				Created:   1236788,
				UserID:    123,
				SpaceID:   uuid.New(),
				Operation: rabbit.CreateOp,
				Text:      "test",
				Type:      model.TextNoteType,
			},
			err: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv := createTestSpaceSrv(t, repo, cache, worker)

			if tt.err == nil {
				worker.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(nil)
			} else {
				worker.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(tt.err)
			}

			err := spaceSrv.CreateNote(context.Background(), tt.req)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetAllNotesBySpaceIDFull(t *testing.T) {
	type test struct {
		name    string
		spaceID uuid.UUID
		want    []model.Note
		err     error
	}

	spaceID := uuid.New()
	tests := []test{
		{
			name:    "positive case",
			spaceID: spaceID,
			want: []model.Note{
				{
					ID:      uuid.New(),
					Created: time.Now(),
					Text:    "test note 1",
					Type:    model.TextNoteType,
				},
				{
					ID:      uuid.New(),
					Created: time.Now(),
					Text:    "test note 2",
					Type:    model.TextNoteType,
				},
			},
			err: nil,
		},
		{
			name:    "error case: db error",
			spaceID: spaceID,
			want:    nil,
			err:     errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv := createTestSpaceSrv(t, repo, cache, worker)

			repo.EXPECT().GetAllNotesBySpaceIDFull(gomock.Any(), gomock.Any()).Return(tt.want, tt.err)

			got, err := spaceSrv.GetAllNotesBySpaceIDFull(context.Background(), tt.spaceID)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetAllNotesBySpaceID(t *testing.T) {
	type test struct {
		name    string
		spaceID uuid.UUID
		want    []model.GetNote
		err     error
	}

	spaceID := uuid.New()
	tests := []test{
		{
			name:    "positive case",
			spaceID: spaceID,
			want: []model.GetNote{
				{
					ID:      uuid.New(),
					Created: time.Now(),
					UserID:  123,
					Text:    "test note 1",
					Type:    model.TextNoteType,
				},
			},
			err: nil,
		},
		{
			name:    "error case: db error",
			spaceID: spaceID,
			want:    nil,
			err:     errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv := createTestSpaceSrv(t, repo, cache, worker)

			repo.EXPECT().GetAllNotesBySpaceID(gomock.Any(), gomock.Any()).Return(tt.want, tt.err)

			got, err := spaceSrv.GetAllNotesBySpaceID(context.Background(), tt.spaceID)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUpdateNote(t *testing.T) {
	type test struct {
		name string
		req  rabbit.UpdateNoteRequest
		err  error
	}

	tests := []test{
		{
			name: "positive case",
			req: rabbit.UpdateNoteRequest{
				ID:        uuid.New(),
				UserID:    123,
				SpaceID:   uuid.New(),
				Operation: rabbit.UpdateOp,
				Text:      "updated test",
			},
			err: nil,
		},
		{
			name: "error case: db error",
			req: rabbit.UpdateNoteRequest{
				ID:        uuid.New(),
				UserID:    123,
				SpaceID:   uuid.New(),
				Operation: rabbit.UpdateOp,
				Text:      "updated test",
			},
			err: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv := createTestSpaceSrv(t, repo, cache, worker)

			worker.EXPECT().UpdateNote(context.Background(), &tt.req).Return(tt.err)

			err := spaceSrv.UpdateNote(context.Background(), tt.req)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetNoteByID(t *testing.T) {
	type test struct {
		name   string
		noteID uuid.UUID
		want   model.GetNote
		err    error
	}

	noteID := uuid.New()
	tests := []test{
		{
			name:   "positive case",
			noteID: noteID,
			want: model.GetNote{
				ID:      noteID,
				Created: time.Now(),
				UserID:  123,
				Text:    "test note",
				Type:    model.TextNoteType,
			},
			err: nil,
		},
		{
			name:   "error case: note not found",
			noteID: noteID,
			want:   model.GetNote{},
			err:    errors.New("note not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv := createTestSpaceSrv(t, repo, cache, worker)

			repo.EXPECT().GetNoteByID(gomock.Any(), gomock.Any()).Return(tt.want, tt.err)

			got, err := spaceSrv.GetNoteByID(context.Background(), tt.noteID)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetNotesTypes(t *testing.T) {
	type test struct {
		name    string
		spaceID uuid.UUID
		want    []model.NoteTypeResponse
		err     error
	}

	spaceID := uuid.New()
	tests := []test{
		{
			name:    "positive case",
			spaceID: spaceID,
			want: []model.NoteTypeResponse{
				{
					Type:  model.TextNoteType,
					Count: 3,
				},
				{
					Type:  model.PhotoNoteType,
					Count: 2,
				},
			},
			err: nil,
		},
		{
			name:    "error case: db error",
			spaceID: spaceID,
			want:    nil,
			err:     errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv := createTestSpaceSrv(t, repo, cache, worker)

			repo.EXPECT().GetNotesTypes(gomock.Any(), gomock.Any()).Return(tt.want, tt.err)

			got, err := spaceSrv.GetNotesTypes(context.Background(), tt.spaceID)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetNotesByType(t *testing.T) {
	type test struct {
		name     string
		spaceID  uuid.UUID
		noteType model.NoteType
		want     []model.GetNote
		err      error
	}

	spaceID := uuid.New()
	tests := []test{
		{
			name:     "positive case",
			spaceID:  spaceID,
			noteType: model.TextNoteType,
			want: []model.GetNote{
				{
					ID:      uuid.New(),
					Created: time.Now(),
					UserID:  123,
					Text:    "test note 1",
					Type:    model.TextNoteType,
				},
				{
					ID:      uuid.New(),
					Created: time.Now(),
					UserID:  123,
					Text:    "test note 2",
					Type:    model.TextNoteType,
				},
			},
			err: nil,
		},
		{
			name:     "error case: db error",
			spaceID:  spaceID,
			noteType: model.TextNoteType,
			want:     nil,
			err:      errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv := createTestSpaceSrv(t, repo, cache, worker)

			repo.EXPECT().GetNotesByType(gomock.Any(), gomock.Any(), gomock.Any()).Return(tt.want, tt.err)

			got, err := spaceSrv.GetNotesByType(context.Background(), tt.spaceID, tt.noteType)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSearchNoteByText(t *testing.T) {
	type test struct {
		name string
		req  model.SearchNoteByTextRequest
		want []model.GetNote
		err  error
	}

	tests := []test{
		{
			name: "positive case",
			req: model.SearchNoteByTextRequest{
				SpaceID: uuid.New(),
				Text:    "test",
				Type:    model.TextNoteType,
			},
			want: []model.GetNote{
				{
					ID:      uuid.New(),
					Created: time.Now(),
					UserID:  123,
					Text:    "test note",
					Type:    model.TextNoteType,
				},
			},
			err: nil,
		},
		{
			name: "positive case: default type",
			req: model.SearchNoteByTextRequest{
				SpaceID: uuid.New(),
				Text:    "test",
			},
			want: []model.GetNote{
				{
					ID:      uuid.New(),
					Created: time.Now(),
					UserID:  123,
					Text:    "test note",
					Type:    model.TextNoteType,
				},
			},
			err: nil,
		},
		{
			name: "error case: db error",
			req: model.SearchNoteByTextRequest{
				SpaceID: uuid.New(),
				Text:    "test",
				Type:    model.TextNoteType,
			},
			want: nil,
			err:  errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv := createTestSpaceSrv(t, repo, cache, worker)

			expectedReq := tt.req
			if len(tt.req.Type) == 0 {
				expectedReq.Type = model.TextNoteType
			}

			repo.EXPECT().SearchNoteByText(gomock.Any(), gomock.Any()).Return(tt.want, tt.err)

			got, err := spaceSrv.SearchNoteByText(context.Background(), tt.req)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestDeleteNote(t *testing.T) {
	type test struct {
		name string
		req  rabbit.DeleteNoteRequest
		err  error
	}

	tests := []test{
		{
			name: "positive case",
			req: rabbit.DeleteNoteRequest{
				ID:        uuid.New(),
				SpaceID:   uuid.New(),
				Operation: rabbit.DeleteOp,
			},
			err: nil,
		},
		{
			name: "error case: db error",
			req: rabbit.DeleteNoteRequest{
				ID:        uuid.New(),
				SpaceID:   uuid.New(),
				Operation: rabbit.DeleteOp,
			},
			err: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv := createTestSpaceSrv(t, repo, cache, worker)

			worker.EXPECT().DeleteNote(context.Background(), &tt.req).Return(tt.err)

			err := spaceSrv.DeleteNote(context.Background(), tt.req)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteAllNotes(t *testing.T) {
	type test struct {
		name string
		req  rabbit.DeleteAllNotesRequest
		err  error
	}

	tests := []test{
		{
			name: "positive case",
			req: rabbit.DeleteAllNotesRequest{
				SpaceID:   uuid.New(),
				Operation: rabbit.DeleteAllOp,
			},
			err: nil,
		},
		{
			name: "error case: db error",
			req: rabbit.DeleteAllNotesRequest{
				SpaceID:   uuid.New(),
				Operation: rabbit.DeleteAllOp,
			},
			err: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, cache, worker := createMockServices(ctrl)
			spaceSrv := createTestSpaceSrv(t, repo, cache, worker)

			worker.EXPECT().DeleteAllNotes(context.Background(), &tt.req).Return(tt.err)

			err := spaceSrv.DeleteAllNotes(context.Background(), tt.req)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
