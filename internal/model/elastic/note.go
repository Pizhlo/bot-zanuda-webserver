package elastic

import (
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/deletebyquery"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/update"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/operator"
	"github.com/google/uuid"
)

// Структура для хранения и поиска по заметкам
type Note struct {
	ID        uuid.UUID // id из базы
	ElasticID string    // id в elastic Search
	TgID      int64
	Text      string
}

// ValidateNote проверяет поля структуры elastic.Data на правильность и возвращает заметку
func (n Note) validate() error {
	return nil
}

func (n Note) getVal() interface{} {
	return n
}

func (n Note) searchByTextQuery() (*search.Request, error) {
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
							"TgID": {
								Query: fmt.Sprintf("%d", n.TgID),
							},
						},
					},
				},
			},
		},
	}

	req := &search.Request{
		Query: &types.Query{
			Bool: &types.BoolQuery{
				Must: must1,
			},
		},
	}

	return req, nil
}

func (n Note) searchByIDQuery() (*search.Request, error) {
	req := &search.Request{
		Query: &types.Query{
			Match: map[string]types.MatchQuery{
				"ID": {
					Query: n.ID.String(),
				},
			},
		},
	}

	return req, nil
}

func (n Note) deleteByQuery() (*deletebyquery.Request, error) {
	req := &deletebyquery.Request{
		Query: &types.Query{
			Match: map[string]types.MatchQuery{
				"TgID": {
					Query: fmt.Sprintf("%d", n.TgID),
				},
			},
		},
	}

	return req, nil
}

func (n Note) updateQuery() (*update.Request, error) {
	data := map[string]string{"Text": n.Text}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req := &update.Request{
		Doc: dataBytes,
	}

	return req, nil
}

func (n *Note) setElasticID(id string) {
	n.ElasticID = id
}

func valueToPointer[T string | float32 | float64](val T) *T {
	return &val
}
