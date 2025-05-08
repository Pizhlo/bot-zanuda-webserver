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
		expected    Config
		err         error
	}

	tests := []test{
		{
			name: "positive case",
			queuesNames: map[string]string{
				CreateNoteQueueNameKey: "q1",
				UpdateNoteQueueNameKey: "q2",
				DeleteNoteQueueNameKey: "q3",
			},
			addr: "amqp://addr",
			expected: Config{
				CreateNoteQueueName: "q1",
				UpdateNoteQueueName: "q2",
				DeleteNoteQueueName: "q3",
				Address:             "amqp://addr",
			},
		},
		{
			name: "addr with prefix amqps",
			queuesNames: map[string]string{
				CreateNoteQueueNameKey: "q1",
				UpdateNoteQueueNameKey: "q2",
				DeleteNoteQueueNameKey: "q3",
			},
			addr: "amqps://addr",
			expected: Config{
				CreateNoteQueueName: "q1",
				UpdateNoteQueueName: "q2",
				DeleteNoteQueueName: "q3",
				Address:             "amqps://addr",
			},
		},
		{
			name: "empty address",
			queuesNames: map[string]string{
				CreateNoteQueueNameKey: "q1",
				UpdateNoteQueueNameKey: "q2",
				DeleteNoteQueueNameKey: "q3",
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
				CreateNoteQueueNameKey: "q1",
				UpdateNoteQueueNameKey: "q2",
				DeleteNoteQueueNameKey: "q3",
			},
			err: errors.New("invalid rabbitMQ address: addr"),
		},
		{
			name: "create note queue name not provided",
			queuesNames: map[string]string{
				UpdateNoteQueueNameKey: "q2",
				DeleteNoteQueueNameKey: "q3",
			},
			addr: "amqps://addr",
			err:  fmt.Errorf("create note queue name not provided"),
		},
		{
			name: "update note queue name not provided",
			queuesNames: map[string]string{
				CreateNoteQueueNameKey: "q2",
				DeleteNoteQueueNameKey: "q3",
			},
			addr: "amqps://addr",
			err:  fmt.Errorf("update note queue name not provided"),
		},
		{
			name: "delete note queue name not provided",
			queuesNames: map[string]string{
				CreateNoteQueueNameKey: "q2",
				UpdateNoteQueueNameKey: "q3",
			},
			addr: "amqps://addr",
			err:  fmt.Errorf("delete note queue name not provided"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := NewConfig(tt.queuesNames, tt.addr)
			assert.Equal(t, tt.expected, actual)

			if tt.err != nil {
				assert.EqualError(t, tt.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
