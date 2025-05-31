package worker

import (
	"context"
	"encoding/json"
	"webserver/internal/model/rabbit"
)

func (s *worker) CreateNote(ctx context.Context, req rabbit.Model) error {
	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.cfg.NotesTopicName, bodyJSON)
}

func (s *worker) UpdateNote(ctx context.Context, req rabbit.Model) error {
	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.cfg.NotesTopicName, bodyJSON)
}

func (s *worker) DeleteNote(ctx context.Context, req rabbit.Model) error {
	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.cfg.NotesTopicName, bodyJSON)
}

func (s *worker) DeleteAllNotes(ctx context.Context, req rabbit.Model) error {
	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.cfg.NotesTopicName, bodyJSON)
}
