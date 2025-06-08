package worker

import (
	"context"
	"encoding/json"
	"webserver/internal/model/rabbit"
)

func (s *Worker) CreateNote(ctx context.Context, req rabbit.Model) error {
	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.config.notesTopic, bodyJSON)
}

func (s *Worker) UpdateNote(ctx context.Context, req rabbit.Model) error {
	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.config.notesTopic, bodyJSON)
}

func (s *Worker) DeleteNote(ctx context.Context, req rabbit.Model) error {
	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.config.notesTopic, bodyJSON)
}

func (s *Worker) DeleteAllNotes(ctx context.Context, req rabbit.Model) error {
	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.config.notesTopic, bodyJSON)
}
