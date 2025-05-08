package worker

import (
	"fmt"
	"strings"
)

type config struct {
	NotesTopicName string
	Address        string
}

const (
	NotesTopicName = "notesTopicName"
)

// NewConfig создает конфиг, принимая на вход мапу с названиями топиков, и адрес сервера.
// Для правильного наименования ключей в мапе следует использовать ключи: CreateNoteQueueNameKey, UpdateNoteQueueNameKey, DeleteNoteQueueNameKey.
func NewConfig(queuesNames map[string]string, addr string) (config, error) {
	if len(addr) == 0 {
		return config{}, fmt.Errorf("address not provided")
	}

	if len(queuesNames) == 0 {
		return config{}, fmt.Errorf("queue names not provided")
	}

	if !strings.HasPrefix(addr, "amqp://") && !strings.HasPrefix(addr, "amqps://") {
		return config{}, fmt.Errorf("invalid rabbitMQ address: %+v", addr)
	}

	cfg := config{
		Address: addr,
	}

	var ok bool
	cfg.NotesTopicName, ok = queuesNames[NotesTopicName]
	if !ok {
		return config{}, fmt.Errorf("notes queue name not provided")
	}

	return cfg, nil
}
