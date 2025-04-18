package model

import (
	"fmt"
	"testing"

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
			name: "SpaceID not filled",
			model: CreateNoteRequest{
				UserID: 1,
				Text:   "test",
			},
			err: fmt.Errorf("invalid space id"),
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
			err:   ErrFieldUserNotFilled,
		},
		{
			name: "text not filled",
			model: Note{
				User: &User{
					TgID: 1,
				},
			},
			err: ErrFieldTextNotFilled,
		},
		{
			name: "Space not filled",
			model: Note{
				User: &User{
					TgID: 1,
				},
				Text: "test",
			},
			err: ErrSpaceIsNil,
		},
		{
			name: "Type not filled",
			model: Note{
				User: &User{
					TgID: 1,
				},
				Space: &Space{},
				Text:  "test",
			},
			err: ErrFieldTypeNotFilled,
		},
		{
			name: "Space filled",
			model: Note{
				User: &User{
					TgID: 1,
				},
				Text:  "test",
				Space: &Space{},
				Type:  TextNoteType,
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
