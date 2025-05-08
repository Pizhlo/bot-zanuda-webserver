package elastic

import (
	"fmt"
	"testing"
	model_package "webserver/internal/model"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/deletebyquery"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/update"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElasticIndexString(t *testing.T) {
	assert.Equal(t, "notes", NoteIndex.String())
	assert.Equal(t, "reminders", ReminderIndex.String())
}

func TestValidateNote(t *testing.T) {
	type test struct {
		name string
		data Data
		err  error
	}

	tests := []test{
		{
			name: "index != NoteIndex",
			data: Data{
				Index: "some another index",
				Model: &Note{
					ID:        uuid.New(),
					ElasticID: "12354",
					TgID:      123,
					Text:      "test",
					SpaceID:   uuid.New(),
					Type:      model_package.TextNoteType,
				},
			},
			err: fmt.Errorf("index is not equal to `notes`: `some another index`"),
		},
		{
			name: "model is not Note",
			data: Data{
				Index: NoteIndex,
				Model: mockNote{},
			},
			err: fmt.Errorf("cannot convert interface{} to elastic.Note. Value: %+v", mockNote{}),
		},
		{
			name: "invalid model",
			data: Data{
				Index: NoteIndex,
				Model: &Note{
					ID:      uuid.New(),
					TgID:    123,
					Text:    "test",
					SpaceID: uuid.New(),
					Type:    model_package.TextNoteType,
				},
			},
			err: ErrFieldElasticIDNotFilled,
		},
		{
			name: "positive case",
			data: Data{
				Index: NoteIndex,
				Model: &Note{
					ID:        uuid.New(),
					TgID:      123,
					ElasticID: "123455",
					Text:      "test",
					SpaceID:   uuid.New(),
					Type:      model_package.TextNoteType,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := tt.data.ValidateNote()
			if tt.err != nil {
				assert.EqualError(t, tt.err, err.Error())
			} else {
				assert.Equal(t, tt.data.Model, actual)
				require.NoError(t, err)
			}
		})
	}
}

type mockNote struct{}

func (mockNote) validate() error {
	return nil
}

func (m mockNote) getVal() any {
	return m
}

func (mockNote) searchByIDQuery() (*search.Request, error) {
	return nil, nil
}

func (mockNote) searchByTextQuery() (*search.Request, error) {
	return nil, nil
}

func (mockNote) deleteByQuery() (*deletebyquery.Request, error) {
	return nil, nil
}

func (mockNote) updateQuery() (*update.Request, error) {
	return nil, nil
}

func (mockNote) setElasticID(s string) {
}
