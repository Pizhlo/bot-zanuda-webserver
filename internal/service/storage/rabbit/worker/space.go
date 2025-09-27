package worker

import (
	"context"
	"encoding/json"
	"webserver/internal/model/rabbit"
)

func (s *Worker) CreateSpace(ctx context.Context, req rabbit.Model) error {
	s.logger.WithField("req", req).Debug("creating space")

	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.config.spacesExchange, rabbit.CreateOp, bodyJSON, req.GetID())
}

func (s *Worker) AddParticipant(ctx context.Context, req rabbit.Model) error {
	s.logger.WithField("req", req).Debug("adding participant")

	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.config.spacesExchange, rabbit.AddParticipantOp, bodyJSON, req.GetID())
}
