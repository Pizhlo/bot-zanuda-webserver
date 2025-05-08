package model

import (
	"testing"
	"time"

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
				UserID:  1,
				Text:    "test",
				SpaceID: uuid.New(),
				Type:    TextNoteType,
				Created: 123,
			},
		},
		{
			name: "user ID not filled",
			model: CreateNoteRequest{
				Text:    "test",
				SpaceID: uuid.New(),
				Type:    TextNoteType,
				Created: 123,
			},
			err: ErrFieldUserNotFilled,
		},
		{
			name: "text not filled",
			model: CreateNoteRequest{
				UserID:  1,
				SpaceID: uuid.New(),
				Type:    TextNoteType,
				Created: 123,
			},
			err: ErrFieldTextNotFilled,
		},
		{
			name: "SpaceID not filled",
			model: CreateNoteRequest{
				UserID:  1,
				Text:    "test",
				Type:    TextNoteType,
				Created: 123,
			},
			err: ErrInvalidSpaceID,
		},
		{
			name: "Type field not filled",
			model: CreateNoteRequest{
				UserID:  1,
				Text:    "test",
				SpaceID: uuid.New(),
				Created: 123,
			},
			err: ErrFieldTypeNotFilled,
		},
		{
			name: "Created field not filled",
			model: CreateNoteRequest{
				UserID:  1,
				Text:    "test",
				SpaceID: uuid.New(),
				Type:    TextNoteType,
			},
			err: ErrFieldCreatedNotFilled,
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

func TestNoteValidate(t *testing.T) {
	type test struct {
		name  string
		model Note
		err   error
	}

	tests := []test{
		{
			name: "positive case",
			model: Note{
				User: &User{
					ID:       1,
					TgID:     123,
					Username: "username",
					PersonalSpace: &Space{
						ID:       uuid.New(),
						Name:     "space1",
						Created:  time.Now(),
						Creator:  123,
						Personal: true,
					},
					Timezone: "Europe/Moscow",
				},
				ID:   uuid.New(),
				Text: "text",
				Space: &Space{
					ID:       uuid.New(),
					Name:     "space2",
					Created:  time.Now(),
					Creator:  234,
					Personal: true,
				},
				Type: TextNoteType,
			},
		},
		{
			name: "field `User` is nil",
			model: Note{
				ID:   uuid.New(),
				Text: "text",
				Space: &Space{
					ID:       uuid.New(),
					Name:     "space2",
					Created:  time.Now(),
					Creator:  234,
					Personal: true,
				},
				Type: TextNoteType,
			},
			err: ErrFieldUserNotFilled,
		},
		{
			name: "text not filled",
			model: Note{
				User: &User{
					ID:       1,
					TgID:     123,
					Username: "username",
					PersonalSpace: &Space{
						ID:       uuid.New(),
						Name:     "space1",
						Created:  time.Now(),
						Creator:  123,
						Personal: true,
					},
					Timezone: "Europe/Moscow",
				},
				ID: uuid.New(),
				Space: &Space{
					ID:       uuid.New(),
					Name:     "space2",
					Created:  time.Now(),
					Creator:  234,
					Personal: true,
				},
				Type: TextNoteType,
			},
			err: ErrFieldTextNotFilled,
		},
		{
			name: "Space not filled",
			model: Note{
				User: &User{
					ID:       1,
					TgID:     123,
					Username: "username",
					PersonalSpace: &Space{
						ID:       uuid.New(),
						Name:     "space1",
						Created:  time.Now(),
						Creator:  123,
						Personal: true,
					},
					Timezone: "Europe/Moscow",
				},
				ID:   uuid.New(),
				Text: "text",
				Type: TextNoteType,
			},
			err: ErrSpaceIsNil,
		},
		{
			name: "Type not filled",
			model: Note{
				User: &User{
					ID:       1,
					TgID:     123,
					Username: "username",
					PersonalSpace: &Space{
						ID:       uuid.New(),
						Name:     "space1",
						Created:  time.Now(),
						Creator:  123,
						Personal: true,
					},
					Timezone: "Europe/Moscow",
				},
				ID:   uuid.New(),
				Text: "text",
				Space: &Space{
					ID:       uuid.New(),
					Name:     "space2",
					Created:  time.Now(),
					Creator:  234,
					Personal: true,
				},
			},
			err: ErrFieldTypeNotFilled,
		},
		{
			name: "Space filled",
			model: Note{
				User: &User{
					ID:       1,
					TgID:     123,
					Username: "username",
					PersonalSpace: &Space{
						ID:       uuid.New(),
						Name:     "space1",
						Created:  time.Now(),
						Creator:  123,
						Personal: true,
					},
					Timezone: "Europe/Moscow",
				},
				Space: &Space{
					ID:       uuid.New(),
					Name:     "space2",
					Created:  time.Now(),
					Personal: true,
					Creator:  123,
				},
				ID:   uuid.New(),
				Text: "text",
				Type: TextNoteType,
			},
			err: nil,
		},
		{
			name: "invalid space",
			model: Note{
				User: &User{
					ID:       1,
					TgID:     123,
					Username: "username",
					PersonalSpace: &Space{
						ID:       uuid.New(),
						Name:     "space1",
						Created:  time.Now(),
						Creator:  123,
						Personal: true,
					},
					Timezone: "Europe/Moscow",
				},
				ID:   uuid.New(),
				Text: "text",
				Space: &Space{
					ID:       uuid.New(),
					Name:     "space2",
					Created:  time.Now(),
					Personal: true,
				},
				Type: TextNoteType,
			},
			err: ErrFieldCreatorNotFilled,
		},
		{
			name: "invalid user",
			model: Note{
				User: &User{
					ID:       1,
					TgID:     123,
					Username: "username",
					Timezone: "Europe/Moscow",
				},
				ID:   uuid.New(),
				Text: "text",
				Space: &Space{
					ID:       uuid.New(),
					Name:     "space2",
					Created:  time.Now(),
					Creator:  234,
					Personal: true,
				},
				Type: TextNoteType,
			},
			err: ErrSpaceIsNil,
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

func TestGetNoteValidate(t *testing.T) {
	type test struct {
		name  string
		model GetNote
		err   error
	}

	tests := []test{
		{
			name: "positive case",
			model: GetNote{
				UserID:  1,
				Text:    "test",
				SpaceID: uuid.New(),
				Type:    TextNoteType,
				Created: time.Now(),
			},
		},
		{
			name: "user ID not filled",
			model: GetNote{
				Text:    "test",
				SpaceID: uuid.New(),
				Type:    TextNoteType,
				Created: time.Now(),
			},
			err: ErrFieldUserNotFilled,
		},
		{
			name: "text not filled",
			model: GetNote{
				UserID:  1,
				SpaceID: uuid.New(),
				Type:    TextNoteType,
				Created: time.Now(),
			},
			err: ErrFieldTextNotFilled,
		},
		{
			name: "SpaceID not filled",
			model: GetNote{
				UserID:  1,
				Text:    "test",
				Type:    TextNoteType,
				Created: time.Now(),
			},
			err: ErrInvalidSpaceID,
		},
		{
			name: "Type field not filled",
			model: GetNote{
				UserID:  1,
				Text:    "test",
				SpaceID: uuid.New(),
				Created: time.Now(),
			},
			err: ErrFieldTypeNotFilled,
		},
		{
			name: "Created field not filled",
			model: GetNote{
				UserID:  1,
				Text:    "test",
				SpaceID: uuid.New(),
				Type:    TextNoteType,
			},
			err: ErrFieldCreatedNotFilled,
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
				UserID:  1,
				Text:    "test",
				SpaceID: uuid.New(),
				Created: 123,
			},
		},
		{
			name: "user ID not filled",
			model: UpdateNoteRequest{
				Text:    "test",
				SpaceID: uuid.New(),
				Created: 123,
			},
			err: ErrFieldUserNotFilled,
		},
		{
			name: "text not filled",
			model: UpdateNoteRequest{
				UserID:  1,
				SpaceID: uuid.New(),
				Created: 123,
			},
			err: ErrFieldTextNotFilled,
		},
		{
			name: "SpaceID not filled",
			model: UpdateNoteRequest{
				UserID:  1,
				Text:    "test",
				Created: 123,
			},
			err: ErrInvalidSpaceID,
		},
		{
			name: "Created field not filled",
			model: UpdateNoteRequest{
				UserID:  1,
				Text:    "test",
				SpaceID: uuid.New(),
			},
			err: ErrFieldCreatedNotFilled,
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
				ID:      uuid.New(),
				SpaceID: uuid.New(),
				NoteID:  uuid.New(),
				Created: 123,
			},
		},
		{
			name: "SpaceID not filled",
			model: DeleteNoteRequest{
				ID:      uuid.New(),
				NoteID:  uuid.New(),
				Created: 123,
			},
			err: ErrInvalidSpaceID,
		},
		{
			name: "ID not filled",
			model: DeleteNoteRequest{
				SpaceID: uuid.New(),
				NoteID:  uuid.New(),
				Created: 123,
			},
			err: ErrFieldIDNotFilled,
		},
		{
			name: "NoteID not filled",
			model: DeleteNoteRequest{
				ID:      uuid.New(),
				SpaceID: uuid.New(),
				Created: 123,
			},
			err: ErrNoteIdNotFilled,
		},
		{
			name: "Created field not filled",
			model: DeleteNoteRequest{
				ID:      uuid.New(),
				SpaceID: uuid.New(),
				NoteID:  uuid.New(),
			},
			err: ErrFieldCreatedNotFilled,
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
