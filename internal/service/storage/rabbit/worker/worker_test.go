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
		want *Worker
		err  error
	}{
		{
			name: "positive case",
			opts: []RabbitOption{
				WithAddress("amqp://localhost:5672/"),
				WithNotesExchange("notes"),
				WithSpacesExchange("spaces"),
			},
			want: &Worker{
				config: struct {
					address        string
					notesExchange  string
					spacesExchange string
				}{
					address:        "amqp://localhost:5672/",
					notesExchange:  "notes",
					spacesExchange: "spaces",
				},
			},
		},
		{
			name: "negative case: address is required",
			opts: []RabbitOption{
				WithNotesExchange("notes"),
				WithSpacesExchange("spaces"),
			},
			err: fmt.Errorf("rabbit: address is required"),
		},
		{
			name: "negative case: notes exchange is required",
			opts: []RabbitOption{
				WithAddress("amqp://localhost:5672/"),
				WithSpacesExchange("spaces"),
			},
			err: fmt.Errorf("rabbit: notes exchange is required"),
		},
		{
			name: "negative case: spaces exchange is required",
			opts: []RabbitOption{
				WithAddress("amqp://localhost:5672/"),
				WithNotesExchange("notes"),
			},
			err: fmt.Errorf("rabbit: spaces exchange is required"),
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
