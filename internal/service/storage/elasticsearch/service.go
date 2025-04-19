package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"webserver/internal/model/elastic"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
)

type client struct {
	addresses []string
	cl        *elasticsearch.TypedClient
}

func New(addresses []string) (*client, error) {
	cl, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: addresses,
	})
	if err != nil {
		return nil, err
	}

	c := &client{cl: cl, addresses: addresses}

	logrus.Infof("sucessfully connected elastic on %v", addresses)

	return c, nil
}

var ErrRecordsNotFound = errors.New(`records not found in elastic`)

// getElasticID ищет запись в elasticSearch по тексту и userID. Возвращает id в elastic search
func (c *client) getElasticID(ctx context.Context, data elastic.Data) (string, error) {
	_, err := data.ValidateNote()
	if err != nil {
		return "", err
	}

	req, err := data.SearchByIDQuery()
	if err != nil {
		return "", err
	}

	res, err := c.cl.Search().
		Index(data.Index.String()).
		Request(req).Do(ctx)
	if err != nil {
		return "", fmt.Errorf("error searching note: %+v", err)
	}

	if len(res.Hits.Hits) == 0 {
		return "", ErrRecordsNotFound
	}

	elasticID := res.Hits.Hits[0].Id_

	return *elasticID, nil
}
