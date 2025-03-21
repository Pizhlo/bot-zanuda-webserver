package worker

import (
	"webserver/internal/model/rabbit"

	amqp "github.com/rabbitmq/amqp091-go"
)

type worker struct {
	conn *amqp.Connection
}

func New(addr string) (*worker, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	return &worker{conn: conn}, nil
}

func (s *worker) Close() error {
	return s.conn.Close()
}

func (s *worker) AddQuery(req rabbit.Request) error {
	return nil
}
