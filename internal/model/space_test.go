package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpaceValidate(t *testing.T) {
	type test struct {
		name  string
		space Space
		err   error
	}

	tests := []test{
		{
			name: "positive case: not personal",
			space: Space{
				ID:       uuid.New(),
				Name:     "space",
				Created:  time.Now().Unix(),
				Creator:  123,
				Personal: false,
			},
		},
		{
			name: "positive case: personal",
			space: Space{
				ID:       uuid.New(),
				Name:     "space",
				Created:  time.Now().Unix(),
				Creator:  123,
				Personal: true,
			},
		},
		{
			name: "uuid nil",
			space: Space{
				Name:     "space",
				Created:  time.Now().Unix(),
				Creator:  123,
				Personal: false,
			},
			err: ErrFieldIDNotFilled,
		},
		{
			name: "empty name",
			space: Space{
				ID:       uuid.New(),
				Created:  time.Now().Unix(),
				Creator:  123,
				Personal: false,
			},
			err: ErrFieldNameNotFilled,
		},
		{
			name: "created not filled",
			space: Space{
				ID:       uuid.New(),
				Name:     "space",
				Creator:  123,
				Personal: false,
			},
			err: ErrFieldCreatedNotFilled,
		},
		{
			name: "creator is empty",
			space: Space{
				ID:       uuid.New(),
				Name:     "space",
				Created:  time.Now().Unix(),
				Personal: false,
			},
			err: ErrFieldCreatorNotFilled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.space.Validate()
			if tt.err != nil {
				assert.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
