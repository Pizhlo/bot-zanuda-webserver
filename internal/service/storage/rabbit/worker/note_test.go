package worker

import (
	"context"
	"encoding/json"
	"testing"
	api_model "webserver/internal/model"
	"webserver/internal/model/rabbit"
	"webserver/internal/service/storage/rabbit/worker/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateNote(t *testing.T) {
	type test struct {
		name string
		req  rabbit.CreateNoteRequest
		err  error
	}

	tests := []test{
		{
			name: "positive case",
			req: rabbit.CreateNoteRequest{
				ID:        uuid.New(),
				UserID:    1234,
				SpaceID:   uuid.New(),
				Text:      "test note",
				Type:      api_model.TextNoteType,
				Created:   5678,
				Operation: rabbit.CreateOp,
			},
		},
		{
			name: "invalid note",
			req: rabbit.CreateNoteRequest{
				ID:      uuid.New(),
				SpaceID: uuid.New(),
				Text:    "test note",
				Type:    api_model.TextNoteType,
				Created: 5678,
			},
			err: api_model.ErrFieldUserNotFilled,
		},
	}

	notesTopicName := "notes"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := mocks.NewMockchannel(ctrl)

	w := worker{
		config: struct {
			address     string
			notesTopic  string
			spacesTopic string
		}{
			address:     "amqp://localhost:5672/",
			notesTopic:  notesTopicName,
			spacesTopic: "spaces",
		},
		channel: ch,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				ch.EXPECT().PublishWithContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) {
						assert.Equal(t, notesTopicName, key)
						assert.False(t, mandatory)
						assert.False(t, immediate)
						assert.Equal(t, "application/json", msg.ContentType)

						actualBody, err := json.Marshal(tt.req)
						require.NoError(t, err)

						assert.Equal(t, actualBody, msg.Body)
					}).Return(nil)
			}

			err := w.CreateNote(context.Background(), &tt.req)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateNote(t *testing.T) {
	type test struct {
		name string
		req  rabbit.UpdateNoteRequest
		err  error
	}

	tests := []test{
		{
			name: "positive case",
			req: rabbit.UpdateNoteRequest{
				ID:        uuid.New(),
				UserID:    1234,
				SpaceID:   uuid.New(),
				Text:      "test note",
				Created:   5678,
				Operation: rabbit.UpdateOp,
			},
		},
		{
			name: "invalid note",
			req: rabbit.UpdateNoteRequest{
				ID:      uuid.New(),
				SpaceID: uuid.New(),
				Text:    "test note",
				Created: 5678,
			},
			err: api_model.ErrFieldUserNotFilled,
		},
	}

	notesTopicName := "notes"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := mocks.NewMockchannel(ctrl)

	w := worker{
		config: struct {
			address     string
			notesTopic  string
			spacesTopic string
		}{
			address:     "amqp://localhost:5672/",
			notesTopic:  notesTopicName,
			spacesTopic: "spaces",
		},
		channel: ch,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				ch.EXPECT().PublishWithContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) {
						assert.Equal(t, notesTopicName, key)
						assert.False(t, mandatory)
						assert.False(t, immediate)
						assert.Equal(t, "application/json", msg.ContentType)

						actualBody, err := json.Marshal(tt.req)
						require.NoError(t, err)

						assert.Equal(t, actualBody, msg.Body)
					}).Return(nil)
			}

			err := w.UpdateNote(context.Background(), &tt.req)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteNote(t *testing.T) {
	type test struct {
		name string
		req  rabbit.DeleteNoteRequest
		err  error
	}

	tests := []test{
		{
			name: "positive case",
			req: rabbit.DeleteNoteRequest{
				ID:        uuid.New(),
				NoteID:    uuid.New(),
				SpaceID:   uuid.New(),
				Created:   5678,
				Operation: rabbit.DeleteOp,
			},
		},
		{
			name: "invalid note",
			req: rabbit.DeleteNoteRequest{
				ID:      uuid.New(),
				SpaceID: uuid.New(),
				Created: 5678,
			},
			err: api_model.ErrIDNotFilled,
		},
	}

	notesTopicName := "notes"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := mocks.NewMockchannel(ctrl)

	w := worker{
		config: struct {
			address     string
			notesTopic  string
			spacesTopic string
		}{
			address:     "amqp://localhost:5672/",
			notesTopic:  notesTopicName,
			spacesTopic: "spaces",
		},
		channel: ch,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				ch.EXPECT().PublishWithContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) {
						assert.Equal(t, notesTopicName, key)
						assert.False(t, mandatory)
						assert.False(t, immediate)
						assert.Equal(t, "application/json", msg.ContentType)

						actualBody, err := json.Marshal(tt.req)
						require.NoError(t, err)

						assert.Equal(t, actualBody, msg.Body)
					}).Return(nil)
			}

			err := w.DeleteNote(context.Background(), &tt.req)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
