package model

import (
	"fmt"
	"testing"
	"time"

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
			name:  "user ID not filled",
			model: CreateNoteRequest{},
			err:   fmt.Errorf("field `user_id` not filled"),
		},
		{
			name: "text not filled",
			model: CreateNoteRequest{
				UserID: 1,
			},
			err: fmt.Errorf("field `text` not filled"),
		},
		{
			name: "created not filled",
			model: CreateNoteRequest{
				UserID: 1,
				Text:   "test",
			},
			err: fmt.Errorf("field `created` not filled"),
		},
		{
			name: "SpaceID not filled",
			model: CreateNoteRequest{
				UserID:  1,
				Text:    "test",
				Created: 123566,
			},
			err: fmt.Errorf("field `space_id` not filled"),
		},
		{
			name: "SpaceID not filled",
			model: CreateNoteRequest{
				UserID:  1,
				Text:    "test",
				Created: 123566,
			},
			err: fmt.Errorf("field `space_id` not filled"),
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
			name:  "field `User` is nil",
			model: Note{},
			err:   fmt.Errorf("field `User` is nil"),
		},
		{
			name: "text not filled",
			model: Note{
				User: &User{
					TgID: 1,
				},
			},
			err: fmt.Errorf("field `text` not filled"),
		},
		{
			name: "created not filled",
			model: Note{
				User: &User{
					TgID: 1,
				},
				Text: "test",
			},
			err: fmt.Errorf("field `created` not filled"),
		},
		{
			name: "Space not filled",
			model: Note{
				User: &User{
					TgID: 1,
				},
				Text:    "test",
				Created: time.Date(2025, 2, 12, 0, 0, 0, 0, time.Local),
			},
			err: fmt.Errorf("field `Space` is nil"),
		},
		{
			name: "Space filled",
			model: Note{
				User: &User{
					TgID: 1,
				},
				Text:    "test",
				Created: time.Date(2025, 2, 12, 0, 0, 0, 0, time.Local),
				Space:   &Space{},
			},
			err: nil,
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
