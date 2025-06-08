package elasticsearch

import (
	"context"
	"fmt"

	"webserver/internal/model/elastic"

	"github.com/sirupsen/logrus"
)

func (c *Client) Save(ctx context.Context, data elastic.Data) error {
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

func (c *Client) UpdateNote(ctx context.Context, data elastic.Data) error {
	_, err := data.ValidateNote()
	if err != nil {
		return fmt.Errorf("error while validating note while updating: %+v", err)
	}

	elasticID, err := c.getElasticID(ctx, data)
	if err != nil {
		return fmt.Errorf("error while getting elasctid ID: %+v", err)
	}

	data.SetElasticID(elasticID)

	req, err := data.UpdateQuery()
	if err != nil {
		return fmt.Errorf("error creating request while updating: %+v", err)
	}

	_, err = c.cl.Update(data.Index.String(), elasticID).Request(req).Do(ctx)
	if err != nil {
		return err
	}

	logrus.Debugf("Elastic: sucecssfully updated user's note")
	return nil
}
