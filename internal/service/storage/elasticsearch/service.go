package elasticsearch

import (
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
