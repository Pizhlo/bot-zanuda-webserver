package worker

import (
	"fmt"
	"strings"
)

type config struct {
	NotesTopicName  string
	SpacesTopicName string
	Address         string
}

const (
	NotesTopicNameKey  = "notesTopicName"
	SpacesTopicNameKey = "spacesTopicName"
)

// NewConfig создает конфиг, принимая на вход мапу с названиями топиков, и адрес сервера.
// Для правильного наименования ключей в мапе следует использовать ключи: NotesTopicNameKey, SpacesTopicNameKey
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
	cfg.NotesTopicName, ok = queuesNames[NotesTopicNameKey]
	if !ok {
		return config{}, fmt.Errorf("notes queue name not provided")
	}

	cfg.SpacesTopicName, ok = queuesNames[SpacesTopicNameKey]
	if !ok {
		return config{}, fmt.Errorf("spaces queue name not provided")
	}

	return cfg, nil
}
