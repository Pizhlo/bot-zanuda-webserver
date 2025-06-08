package worker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts []RabbitOption
		want *worker
		err  error
	}{
		{
			name: "positive case",
			opts: []RabbitOption{
				WithAddress("amqp://localhost:5672/"),
				WithNotesTopic("notes"),
				WithSpacesTopic("spaces"),
			},
			want: &worker{
				config: struct {
					address     string
					notesTopic  string
					spacesTopic string
				}{
					address:     "amqp://localhost:5672/",
					notesTopic:  "notes",
					spacesTopic: "spaces",
				},
			},
		},
		{
			name: "negative case: address is required",
			opts: []RabbitOption{
				WithNotesTopic("notes"),
				WithSpacesTopic("spaces"),
			},
			err: fmt.Errorf("rabbit: address is required"),
		},
		{
			name: "negative case: notes topic is required",
			opts: []RabbitOption{
				WithAddress("amqp://localhost:5672/"),
				WithSpacesTopic("spaces"),
			},
			err: fmt.Errorf("rabbit: notes topic is required"),
		},
		{
			name: "negative case: spaces topic is required",
			opts: []RabbitOption{
				WithAddress("amqp://localhost:5672/"),
				WithNotesTopic("notes"),
			},
			err: fmt.Errorf("rabbit: spaces topic is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.opts...)
			if tt.err != nil {
				require.Error(t, err)
				require.Equal(t, tt.err, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
