package worker

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	type test struct {
		name        string
		queuesNames map[string]string
		addr        string
		expected    config
		err         error
	}

	tests := []test{
		{
			name: "positive case",
			queuesNames: map[string]string{
				NotesTopicName: "q1",
			},
			addr: "amqp://addr",
			expected: config{
				NotesTopicName: "q1",
				Address:        "amqp://addr",
			},
		},
		{
			name: "addr with prefix amqps",
			queuesNames: map[string]string{
				NotesTopicName: "q1",
			},
			addr: "amqps://addr",
			expected: config{
				NotesTopicName: "q1",
				Address:        "amqps://addr",
			},
		},
		{
			name: "empty address",
			queuesNames: map[string]string{
				NotesTopicName: "q1",
			},
			err: errors.New("address not provided"),
		},
		{
			name: "empty queues names",
			addr: "amqp://addr",
			err:  errors.New("queue names not provided"),
		},
		{
			name: "invalid address",
			addr: "addr",
			queuesNames: map[string]string{
				NotesTopicName: "q1",
			},
			err: errors.New("invalid rabbitMQ address: addr"),
		},
		{
			name: "notes queue name not provided",
			queuesNames: map[string]string{
				"another key": "q2",
			},
			addr: "amqps://addr",
			err:  fmt.Errorf("notes queue name not provided"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := NewConfig(tt.queuesNames, tt.addr)
			assert.Equal(t, tt.expected, actual)

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
