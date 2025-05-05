package worker

import (
	"context"
	"encoding/json"
	"webserver/internal/model"
)

func (s *worker) CreateNote(ctx context.Context, req model.CreateNoteRequest) error {
	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.cfg.CreateNoteQueueName, bodyJSON)
}

func (s *worker) UpdateNote(ctx context.Context, req model.UpdateNoteRequest) error {
	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.cfg.UpdateNoteQueueName, bodyJSON)
}

func (s *worker) DeleteNote(ctx context.Context, req model.DeleteNoteRequest) error {
	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(ctx, s.cfg.DeleteNoteQueueName, bodyJSON)
}
