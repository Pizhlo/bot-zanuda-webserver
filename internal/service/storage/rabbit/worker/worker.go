package worker

import (
	"context"
	"fmt"
	"webserver/internal/model/rabbit"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	config struct {
		address string

		// exchanges
		notesExchange  string
		spacesExchange string
	}

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

func WithNotesExchange(notesExchange string) RabbitOption {
	return func(w *Worker) {
		w.config.notesExchange = notesExchange
	}
}

func WithSpacesExchange(spacesExchange string) RabbitOption {
	return func(w *Worker) {
		w.config.spacesExchange = spacesExchange
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

	if w.config.notesExchange == "" {
		return nil, fmt.Errorf("rabbit: notes exchange is required")
	}

	if w.config.spacesExchange == "" {
		return nil, fmt.Errorf("rabbit: spaces exchange is required")
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

	// Create notes exchange
	err = ch.ExchangeDeclare(
		s.config.notesExchange, // name
		"topic",                // type
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		return fmt.Errorf("error creating exchange %s: %+v", s.config.notesExchange, err)
	}

	// Create spaces exchange
	err = ch.ExchangeDeclare(
		s.config.spacesExchange, // name
		"topic",                 // type
		true,                    // durable
		false,                   // auto-deleted
		false,                   // internal
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		return fmt.Errorf("error creating exchange %s: %+v", s.config.spacesExchange, err)
	}

	return nil
}

func (s *Worker) Close() error {
	err := s.channel.Close()
	if err != nil {
		logrus.Errorf("worker: error closing channel rabbit mq: %+v", err)
	}

	return s.conn.Close()
}

func (s *Worker) publish(ctx context.Context, exchange string, operation rabbit.Operation, body []byte) error {
	logrus.Debugf("rabbit: publishing message to exchange '%s' with operation '%s': %+v", exchange, operation, string(body))

	return s.channel.PublishWithContext(
		ctx,
		exchange,          // exchange
		string(operation), // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
