package worker

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	config struct {
		address string

		// queues
		notesTopic  string
		spacesTopic string
	}

	// queues
	notesTopic  amqp.Queue
	spacesTopic amqp.Queue

	conn    *amqp.Connection
	channel channel
}

//go:generate mockgen -source ./worker.go -destination=./mocks/rabbit.go -package=mocks
type channel interface {
	PublishWithContext(_ context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error
	Close() error
}

type RabbitOption func(*Worker)

func WithAddress(address string) RabbitOption {
	return func(w *Worker) {
		w.config.address = address
	}
}

func WithNotesTopic(notesTopic string) RabbitOption {
	return func(w *Worker) {
		w.config.notesTopic = notesTopic
	}
}

func WithSpacesTopic(spacesTopic string) RabbitOption {
	return func(w *Worker) {
		w.config.spacesTopic = spacesTopic
	}
}

func New(opts ...RabbitOption) (*Worker, error) {
	w := &Worker{}

	for _, opt := range opts {
		opt(w)
	}

	if w.config.address == "" {
		return nil, fmt.Errorf("rabbit: address is required")
	}

	if w.config.notesTopic == "" {
		return nil, fmt.Errorf("rabbit: notes topic is required")
	}

	if w.config.spacesTopic == "" {
		return nil, fmt.Errorf("rabbit: spaces topic is required")
	}

	return w, nil
}

func (s *Worker) Connect() error {
	conn, err := amqp.Dial(s.config.address)
	if err != nil {
		return err
	}

	s.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	s.channel = ch

	notesTopic, err := ch.QueueDeclare(
		s.config.notesTopic, // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return fmt.Errorf("error creating queue %s: %+v", s.config.notesTopic, err)
	}

	s.notesTopic = notesTopic

	spacesTopic, err := ch.QueueDeclare(
		s.config.spacesTopic, // name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		return fmt.Errorf("error creating queue %s: %+v", s.config.spacesTopic, err)
	}

	s.spacesTopic = spacesTopic

	return nil
}

func (s *Worker) Close() error {
	err := s.channel.Close()
	if err != nil {
		logrus.Errorf("worker: error closing channel rabbit mq: %+v", err)
	}

	return s.conn.Close()
}

func (s *Worker) publish(ctx context.Context, queue string, body []byte) error {
	logrus.Debugf("rabbit: publishing message to queue '%s': %+v", queue, string(body))

	return s.channel.PublishWithContext(
		ctx,
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
