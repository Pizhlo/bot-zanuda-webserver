package worker

import (
	"encoding/json"
	"webserver/internal/model/rabbit"
)

func (s *worker) CreateNote(req rabbit.Request) error {
	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(createNoteQueueName, bodyJSON)
}

func (s *worker) UpdateNote(req rabbit.Request) error {
	bodyJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return s.publish(updateNoteQueueName, bodyJSON)
}
