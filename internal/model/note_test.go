package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				Created: 1236788,
			},
		},
		{
			name: "user ID not filled",
			model: GetNote{
				Text:    "test",
				SpaceID: uuid.New(),
				Type:    TextNoteType,
				Created: 1236788,
			},
			err: ErrFieldUserNotFilled,
		},
		{
			name: "text not filled",
			model: GetNote{
				UserID:  1,
				SpaceID: uuid.New(),
				Type:    TextNoteType,
				Created: 1236788,
			},
			err: ErrFieldTextNotFilled,
		},
		{
			name: "SpaceID not filled",
			model: GetNote{
				UserID:  1,
				Text:    "test",
				Type:    TextNoteType,
				Created: 1236788,
			},
			err: ErrInvalidSpaceID,
		},
		{
			name: "Type field not filled",
			model: GetNote{
				UserID:  1,
				Text:    "test",
				SpaceID: uuid.New(),
				Created: 1236788,
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
