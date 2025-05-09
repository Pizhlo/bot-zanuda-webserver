package worker

import (
	"context"
	"encoding/json"
	"webserver/internal/model/rabbit"
)

func (s *worker) CreateSpace(ctx context.Context, req rabbit.CreateSpaceRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.cfg.SpacesTopicName, bodyJSON)
}
