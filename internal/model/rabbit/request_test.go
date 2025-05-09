package rabbit

import (
	"testing"
	"webserver/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateNoteRequestValidate(t *testing.T) {
	type test struct {
		name  string
		model CreateNoteRequest
		err   error
	}

	tests := []test{
		{
			name: "positive case",
			model: CreateNoteRequest{
				UserID:    1,
				Text:      "test",
				SpaceID:   uuid.New(),
				Type:      model.TextNoteType,
				Created:   123,
				Operation: CreateOp,
			},
		},
		{
			name: "user ID not filled",
			model: CreateNoteRequest{
				Text:      "test",
				SpaceID:   uuid.New(),
				Type:      model.TextNoteType,
				Created:   123,
				Operation: CreateOp,
			},
			err: model.ErrFieldUserNotFilled,
		},
		{
			name: "text not filled",
			model: CreateNoteRequest{
				UserID:    1,
				SpaceID:   uuid.New(),
				Type:      model.TextNoteType,
				Created:   123,
				Operation: CreateOp,
			},
			err: model.ErrFieldTextNotFilled,
		},
		{
			name: "SpaceID not filled",
			model: CreateNoteRequest{
				UserID:    1,
				Text:      "test",
				Type:      model.TextNoteType,
				Created:   123,
				Operation: CreateOp,
			},
			err: model.ErrInvalidSpaceID,
		},
		{
			name: "Type field not filled",
			model: CreateNoteRequest{
				UserID:    1,
				Text:      "test",
				SpaceID:   uuid.New(),
				Created:   123,
				Operation: CreateOp,
			},
			err: model.ErrFieldTypeNotFilled,
		},
		{
			name: "Created field not filled",
			model: CreateNoteRequest{
				UserID:    1,
				Text:      "test",
				SpaceID:   uuid.New(),
				Type:      model.TextNoteType,
				Operation: CreateOp,
			},
			err: model.ErrFieldCreatedNotFilled,
		},
		{
			name: "operation not filled",
			model: CreateNoteRequest{
				UserID:  1,
				Text:    "test",
				SpaceID: uuid.New(),
				Type:    model.TextNoteType,
				Created: 123,
			},
			err: ErrInvalidOperation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			if tt.err != nil {
				assert.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateNoteRequestValidate(t *testing.T) {
	type test struct {
		name  string
		model UpdateNoteRequest
		err   error
	}

	tests := []test{
		{
			name: "positive case",
			model: UpdateNoteRequest{
				UserID:    1,
				Text:      "test",
				SpaceID:   uuid.New(),
				Created:   123,
				Operation: UpdateOp,
			},
		},
		{
			name: "user ID not filled",
			model: UpdateNoteRequest{
				Text:      "test",
				SpaceID:   uuid.New(),
				Created:   123,
				Operation: UpdateOp,
			},
			err: model.ErrFieldUserNotFilled,
		},
		{
			name: "text not filled",
			model: UpdateNoteRequest{
				UserID:    1,
				SpaceID:   uuid.New(),
				Created:   123,
				Operation: UpdateOp,
			},
			err: model.ErrFieldTextNotFilled,
		},
		{
			name: "SpaceID not filled",
			model: UpdateNoteRequest{
				UserID:    1,
				Text:      "test",
				Created:   123,
				Operation: UpdateOp,
			},
			err: model.ErrInvalidSpaceID,
		},
		{
			name: "Created field not filled",
			model: UpdateNoteRequest{
				UserID:    1,
				Text:      "test",
				SpaceID:   uuid.New(),
				Operation: UpdateOp,
			},
			err: model.ErrFieldCreatedNotFilled,
		},
		{
			name: "operation not filled",
			model: UpdateNoteRequest{
				UserID:  1,
				Text:    "test",
				SpaceID: uuid.New(),
				Created: 123,
			},
			err: ErrInvalidOperation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			if tt.err != nil {
				assert.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteNoteRequestValidate(t *testing.T) {
	type test struct {
		name  string
		model DeleteNoteRequest
		err   error
	}

	tests := []test{
		{
			name: "positive case",
			model: DeleteNoteRequest{
				ID:        uuid.New(),
				SpaceID:   uuid.New(),
				NoteID:    uuid.New(),
				Created:   123,
				Operation: DeleteOp,
			},
		},
		{
			name: "SpaceID not filled",
			model: DeleteNoteRequest{
				ID:        uuid.New(),
				NoteID:    uuid.New(),
				Created:   123,
				Operation: DeleteOp,
			},
			err: model.ErrInvalidSpaceID,
		},
		{
			name: "ID not filled",
			model: DeleteNoteRequest{
				SpaceID:   uuid.New(),
				NoteID:    uuid.New(),
				Created:   123,
				Operation: DeleteOp,
			},
			err: model.ErrFieldIDNotFilled,
		},
		{
			name: "NoteID not filled",
			model: DeleteNoteRequest{
				ID:        uuid.New(),
				SpaceID:   uuid.New(),
				Created:   123,
				Operation: DeleteOp,
			},
			err: model.ErrNoteIdNotFilled,
		},
		{
			name: "Created field not filled",
			model: DeleteNoteRequest{
				ID:        uuid.New(),
				SpaceID:   uuid.New(),
				NoteID:    uuid.New(),
				Operation: DeleteOp,
			},
			err: model.ErrFieldCreatedNotFilled,
		},
		{
			name: "operation not filled",
			model: DeleteNoteRequest{
				ID:        uuid.New(),
				SpaceID:   uuid.New(),
				NoteID:    uuid.New(),
				Created:   123,
				Operation: CreateOp,
			},
			err: ErrInvalidOperation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			if tt.err != nil {
				assert.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteAllNotesRequestValidate(t *testing.T) {
	type test struct {
		name  string
		model DeleteAllNotesRequest
		err   error
	}

	tests := []test{
		{
			name: "positive case",
			model: DeleteAllNotesRequest{
				ID:        uuid.New(),
				SpaceID:   uuid.New(),
				Created:   123,
				Operation: DeleteAllOp,
			},
		},
		{
			name: "SpaceID not filled",
			model: DeleteAllNotesRequest{
				ID:        uuid.New(),
				Created:   123,
				Operation: DeleteAllOp,
			},
			err: model.ErrInvalidSpaceID,
		},
		{
			name: "ID not filled",
			model: DeleteAllNotesRequest{
				SpaceID:   uuid.New(),
				Created:   123,
				Operation: DeleteAllOp,
			},
			err: model.ErrFieldIDNotFilled,
		},
		{
			name: "Created field not filled",
			model: DeleteAllNotesRequest{
				ID:        uuid.New(),
				SpaceID:   uuid.New(),
				Operation: DeleteAllOp,
			},
			err: model.ErrFieldCreatedNotFilled,
		},
		{
			name: "operation not filled",
			model: DeleteAllNotesRequest{
				ID:        uuid.New(),
				SpaceID:   uuid.New(),
				Created:   123,
				Operation: CreateOp,
			},
			err: ErrInvalidOperation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			if tt.err != nil {
				assert.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
