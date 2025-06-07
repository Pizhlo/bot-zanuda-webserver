package worker

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type worker struct {
	cfg config

	conn    *amqp.Connection
	channel channel

	// queues
	notesTopic  amqp.Queue
	spacesTopic amqp.Queue
}

//go:generate mockgen -source ./worker.go -destination=./mocks/rabbit.go -package=mocks
type channel interface {
	PublishWithContext(_ context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error
	Close() error
}

func New(cfg config) *worker {
	return &worker{
		cfg: cfg,
	}
}

func (s *worker) Connect() error {
	conn, err := amqp.Dial(s.cfg.Address)
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
		s.cfg.NotesTopicName, // name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		return fmt.Errorf("error creating queue %s: %+v", s.cfg.NotesTopicName, err)
	}

	s.notesTopic = notesTopic

	spacesTopic, err := ch.QueueDeclare(
		s.cfg.SpacesTopicName, // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		return fmt.Errorf("error creating queue %s: %+v", s.cfg.SpacesTopicName, err)
	}

	s.spacesTopic = spacesTopic

	return nil
}

func (s *worker) Close() error {
	err := s.channel.Close()
	if err != nil {
		logrus.Errorf("worker: error closing channel rabbit mq: %+v", err)
	}

	return s.conn.Close()
}

func (s *worker) publish(ctx context.Context, queue string, body []byte) error {
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
