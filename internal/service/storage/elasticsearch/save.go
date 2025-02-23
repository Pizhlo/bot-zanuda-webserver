package elasticsearch

import (
	"context"
	"fmt"

	"webserver/internal/model/elastic"

	"github.com/sirupsen/logrus"
)

func (c *client) Save(ctx context.Context, data elastic.Data) error {
	_, err := data.ValidateNote()
	if err != nil {
		return fmt.Errorf("error validating note while saving: %+v", err)
	}

	_, err = c.cl.Index(data.Index.String()).
		Request(data.Model).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("error while saving note in elastic: %+v", err)
	}

	logrus.Debugf("Elastic: sucecssfully saved user's note")

	return nil
}
