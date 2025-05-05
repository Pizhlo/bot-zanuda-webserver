package worker

import (
	"fmt"
	"strings"
)

type Config struct {
	CreateNoteQueueName string
	UpdateNoteQueueName string
	DeleteNoteQueueName string
	Address             string
}

const (
	CreateNoteQueueNameKey = "createNoteQueueName"
	UpdateNoteQueueNameKey = "updateNoteQueueName"
	DeleteNoteQueueNameKey = "deleteNoteQueueName"
)

// NewConfig создает конфиг, принимая на вход мапу с названиями топиков, и адрес сервера.
// Для правильного наименования ключей в мапе следует использовать ключи: CreateNoteQueueNameKey, UpdateNoteQueueNameKey, DeleteNoteQueueNameKey.
func NewConfig(queuesNames map[string]string, addr string) (Config, error) {
	if len(addr) == 0 {
		return Config{}, fmt.Errorf("address not provided")
	}

	if len(queuesNames) == 0 {
		return Config{}, fmt.Errorf("queue names not provided")
	}

	if !strings.HasPrefix(addr, "amqp://") && !strings.HasPrefix(addr, "amqps://") {
		return Config{}, fmt.Errorf("invalid rabbitMQ address: %+v", addr)
	}

	cfg := Config{
		Address: addr,
	}

	var ok bool
	cfg.CreateNoteQueueName, ok = queuesNames[CreateNoteQueueNameKey]
	if !ok {
		return Config{}, fmt.Errorf("create note queue name not provided")
	}

	cfg.UpdateNoteQueueName, ok = queuesNames[UpdateNoteQueueNameKey]
	if !ok {
		return Config{}, fmt.Errorf("update note queue name not provided")
	}

	cfg.DeleteNoteQueueName, ok = queuesNames[DeleteNoteQueueNameKey]
	if !ok {
		return Config{}, fmt.Errorf("delete note queue name not provided")
	}

	return cfg, nil
}
