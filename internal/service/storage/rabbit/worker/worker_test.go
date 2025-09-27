package worker

import (
	"fmt"
	"testing"

	"github.com/ex-rate/logger"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	logger, err := logger.New(logger.Config{
		Level:  logger.DebugLevel,
		Output: logger.ConsoleOutput,
	})
	require.NoError(t, err)

	workerLogger := logger.WithService("worker")

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
				WithLogger(workerLogger),
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
				logger: workerLogger,
			},
		},
		{
			name: "negative case: address is required",
			opts: []RabbitOption{
				WithNotesExchange("notes"),
				WithSpacesExchange("spaces"),
				WithLogger(workerLogger),
			},
			err: fmt.Errorf("rabbit: address is required"),
		},
		{
			name: "negative case: notes exchange is required",
			opts: []RabbitOption{
				WithAddress("amqp://localhost:5672/"),
				WithSpacesExchange("spaces"),
				WithLogger(workerLogger),
			},
			err: fmt.Errorf("rabbit: notes exchange is required"),
		},
		{
			name: "negative case: spaces exchange is required",
			opts: []RabbitOption{
				WithAddress("amqp://localhost:5672/"),
				WithNotesExchange("notes"),
				WithLogger(workerLogger),
			},
			err: fmt.Errorf("rabbit: spaces exchange is required"),
		},
		{
			name: "negative case: logger is required",
			opts: []RabbitOption{
				WithAddress("amqp://localhost:5672/"),
				WithNotesExchange("notes"),
				WithSpacesExchange("spaces"),
			},
			err: fmt.Errorf("rabbit: logger is required"),
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
