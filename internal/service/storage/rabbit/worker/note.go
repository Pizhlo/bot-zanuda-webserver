package worker

import (
	"context"
	"encoding/json"
	"webserver/internal/model/rabbit"
)

func (s *Worker) CreateNote(ctx context.Context, req rabbit.Model) error {
	s.logger.WithField("request_id", req.GetID()).Debug("creating note")

	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.config.notesExchange, rabbit.CreateOp, bodyJSON, req.GetID())
}

func (s *Worker) UpdateNote(ctx context.Context, req rabbit.Model) error {
	s.logger.WithField("request_id", req.GetID()).Debug("updating note")

	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.config.notesExchange, rabbit.UpdateOp, bodyJSON, req.GetID())
}

func (s *Worker) DeleteNote(ctx context.Context, req rabbit.Model) error {
	s.logger.WithField("request_id", req.GetID()).Debug("deleting note")

	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.config.notesExchange, rabbit.DeleteOp, bodyJSON, req.GetID())
}

func (s *Worker) DeleteAllNotes(ctx context.Context, req rabbit.Model) error {
	s.logger.WithField("request_id", req.GetID()).Debug("deleting all notes")

	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.config.notesExchange, rabbit.DeleteAllOp, bodyJSON, req.GetID())
}
