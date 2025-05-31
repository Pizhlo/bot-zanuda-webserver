package worker

import (
	"context"
	"encoding/json"
)

func (s *worker) CreateNote(ctx context.Context, req Model) error {
	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.cfg.NotesTopicName, bodyJSON)
}

func (s *worker) UpdateNote(ctx context.Context, req Model) error {
	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.cfg.NotesTopicName, bodyJSON)
}

func (s *worker) DeleteNote(ctx context.Context, req Model) error {
	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.cfg.NotesTopicName, bodyJSON)
}

func (s *worker) DeleteAllNotes(ctx context.Context, req Model) error {
	if err := req.Validate(); err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.cfg.NotesTopicName, bodyJSON)
}
