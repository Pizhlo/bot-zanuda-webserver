package elastic

import (
	"encoding/json"
	"fmt"
	"testing"
	model_package "webserver/internal/model"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/deletebyquery"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/update"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/operator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoteValidate(t *testing.T) {
	type test struct {
		name  string
		model Note
		err   error
	}

	tests := []test{
		{
			name: "positive case",
			model: Note{
				ID:        uuid.New(),
				ElasticID: "12354",
				TgID:      123,
				Text:      "test",
				SpaceID:   uuid.New(),
				Type:      model_package.TextNoteType,
			},
		},
		{
			name: "empty ID",
			model: Note{
				ElasticID: "12354",
				TgID:      123,
				Text:      "test",
				SpaceID:   uuid.New(),
				Type:      model_package.TextNoteType,
			},
			err: ErrFieldIDNotFilled,
		},
		{
			name: "empty tgID",
			model: Note{
				ID:        uuid.New(),
				ElasticID: "12354",
				Text:      "test",
				SpaceID:   uuid.New(),
				Type:      model_package.TextNoteType,
			},
			err: ErrTgIDNotFilled,
		},
		{
			name: "empty elastic id",
			model: Note{
				ID:      uuid.New(),
				TgID:    123,
				Text:    "test",
				SpaceID: uuid.New(),
				Type:    model_package.TextNoteType,
			},
			err: ErrFieldElasticIDNotFilled,
		},
		{
			name: "empty text: type text",
			model: Note{
				ID:        uuid.New(),
				ElasticID: "12354",
				TgID:      123,
				SpaceID:   uuid.New(),
				Type:      model_package.TextNoteType,
			},
			err: ErrFieldTextNotFilled,
		},
		{
			name: "empty text: type photo",
			model: Note{
				ID:        uuid.New(),
				ElasticID: "12354",
				TgID:      123,
				SpaceID:   uuid.New(),
				Type:      model_package.PhotoNoteType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.validate()
			if tt.err != nil {
				assert.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSearchByTextQuery(t *testing.T) {
	n := Note{
		ID:        uuid.New(),
		ElasticID: "123456",
		TgID:      12345,
		Text:      "text note",
		SpaceID:   uuid.New(),
	}

	must1 := []types.Query{
		{
			Bool: &types.BoolQuery{
				Should: []types.Query{
					{
						Match: map[string]types.MatchQuery{
							"Text": {
								Query:     n.Text,
								Operator:  &operator.Or,
								Fuzziness: "auto",
							},
						},
					},
					{
						Wildcard: map[string]types.WildcardQuery{
							"Text": {
								Value:   valueToPointer(fmt.Sprintf("*%s*", n.Text)),
								Boost:   valueToPointer(float32(1.0)),
								Rewrite: valueToPointer("constant_score"),
							},
						},
					},
				},
			},
		},
		{
			Bool: &types.BoolQuery{
				Must: []types.Query{
					{
						Match: map[string]types.MatchQuery{
							"SpaceID": {
								Query: n.SpaceID.String(),
							},
						},
					},
				},
			},
		},
		{
			Bool: &types.BoolQuery{
				Must: []types.Query{
					{
						Match: map[string]types.MatchQuery{
							"Type": {
								Query: string(n.Type),
							},
						},
					},
				},
			},
		},
	}

	result := &search.Request{
		Query: &types.Query{
			Bool: &types.BoolQuery{
				Must: must1,
			},
		},
	}

	actual, err := n.searchByTextQuery()
	require.NoError(t, err)

	assert.Equal(t, result, actual)
}

func TestGetVal(t *testing.T) {
	n := Note{
		ID:        uuid.New(),
		ElasticID: "123456",
		TgID:      12345,
		Text:      "text note",
		SpaceID:   uuid.New(),
	}

	assert.Equal(t, n, n.getVal())
}

func TestSearchByIDQuery(t *testing.T) {
	n := Note{
		ID:        uuid.New(),
		ElasticID: "123456",
		TgID:      12345,
		Text:      "text note",
		SpaceID:   uuid.New(),
	}

	result := &search.Request{
		Query: &types.Query{
			Match: map[string]types.MatchQuery{
				"ID": {
					Query: n.ID.String(),
				},
			},
		},
	}

	actual, err := n.searchByIDQuery()
	require.NoError(t, err)

	assert.Equal(t, result, actual)
}

func TestDeleteByQuery(t *testing.T) {
	n := Note{
		ID:        uuid.New(),
		ElasticID: "123456",
		TgID:      12345,
		Text:      "text note",
		SpaceID:   uuid.New(),
	}

	result := &deletebyquery.Request{
		Query: &types.Query{
			Match: map[string]types.MatchQuery{
				"TgID": {
					Query: fmt.Sprintf("%d", n.TgID),
				},
			},
		},
	}

	actual, err := n.deleteByQuery()
	require.NoError(t, err)

	assert.Equal(t, result, actual)
}

func TestUpdateByQuery(t *testing.T) {
	n := Note{
		ID:        uuid.New(),
		ElasticID: "123456",
		TgID:      12345,
		Text:      "text note",
		SpaceID:   uuid.New(),
	}

	data := map[string]string{"Text": n.Text}

	dataBytes, err := json.Marshal(data)
	require.NoError(t, err)

	result := &update.Request{
		Doc: dataBytes,
	}

	actual, err := n.updateQuery()
	require.NoError(t, err)

	assert.Equal(t, result, actual)
}

func TestSetElasticID(t *testing.T) {
	n := Note{
		ID:      uuid.New(),
		TgID:    12345,
		Text:    "text note",
		SpaceID: uuid.New(),
	}

	elasticID := "88888"

	n.setElasticID(elasticID)

	assert.Equal(t, elasticID, n.ElasticID)
}
