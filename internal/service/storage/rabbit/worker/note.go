package worker

import (
	"encoding/json"
	"webserver/internal/model"
)

func (s *worker) CreateNote(req model.CreateNoteRequest) error {
	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(createNoteQueueName, bodyJSON)
}

func (s *worker) UpdateNote(req model.UpdateNoteRequest) error {
	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(updateNoteQueueName, bodyJSON)
}
