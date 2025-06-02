package rabbit

import (
	"testing"
	"webserver/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSpaceRequestValidate(t *testing.T) {
	type test struct {
		name  string
		model CreateSpaceRequest
		err   error
	}

	tests := []test{
		{
			name: "positive case",
			model: CreateSpaceRequest{
				ID:        uuid.New(),
				UserID:    1,
				Name:      "test name",
				Created:   123,
				Operation: CreateOp,
			},
		},
		{
			name: "ID not filled",
			model: CreateSpaceRequest{
				UserID:    1,
				Name:      "test name",
				Created:   123,
				Operation: CreateOp,
			},
			err: model.ErrIDNotFilled,
		},
		{
			name: "user ID not filled",
			model: CreateSpaceRequest{
				ID:        uuid.New(),
				Name:      "test name",
				Created:   123,
				Operation: CreateOp,
			},
			err: model.ErrFieldUserNotFilled,
		},
		{
			name: "name not filled",
			model: CreateSpaceRequest{
				ID:        uuid.New(),
				UserID:    1,
				Created:   123,
				Operation: CreateOp,
			},
			err: model.ErrFieldNameNotFilled,
		},
		{
			name: "Created field not filled",
			model: CreateSpaceRequest{
				ID:        uuid.New(),
				UserID:    1,
				Name:      "test name",
				Operation: CreateOp,
			},
			err: model.ErrFieldCreatedNotFilled,
		},
		{
			name: "operation not filled",
			model: CreateSpaceRequest{
				ID:        uuid.New(),
				UserID:    1,
				Name:      "test name",
				Created:   123,
				Operation: UpdateOp,
			},
			err: ErrInvalidOperation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
