package dbsaver

import "webserver/internal/model/rabbit"

// saver -сущность, которая отправляет запрос на сохранение / обновление записей в БД.
// отправляет запросы в очередь rabbitMQ
type saver struct {
	queue queue
}

// queue - очередь из запросов в rabbitMQ
type queue interface {
	AddQuery(req rabbit.Request) error
}

func New(queue queue) *saver {
	return &saver{queue: queue}
}

func (s *saver) AddQuery(req rabbit.Request) error {
	return s.queue.AddQuery(req)
}
