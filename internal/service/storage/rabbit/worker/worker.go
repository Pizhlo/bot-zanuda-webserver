package worker

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type worker struct {
	cfg Config

	conn    *amqp.Connection
	channel *amqp.Channel

	// queues
	createNoteQueue amqp.Queue
	updateNoteQueue amqp.Queue
	deleteNoteQueue amqp.Queue
}

// const (
// 	createNoteQueueName = "create_note"
// 	updateNoteQueueName = "update_note"
// )

func New(cfg Config) *worker {
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

	createNoteQueue, err := ch.QueueDeclare(
		s.cfg.CreateNoteQueueName, // name
		true,                      // durable
		false,                     // delete when unused
		false,                     // exclusive
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		return fmt.Errorf("error creating queue %s: %+v", s.cfg.CreateNoteQueueName, err)
	}

	s.createNoteQueue = createNoteQueue

	updateNoteQueue, err := ch.QueueDeclare(
		s.cfg.UpdateNoteQueueName, // name
		true,                      // durable
		false,                     // delete when unused
		false,                     // exclusive
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		return fmt.Errorf("error creating queue %s: %+v", s.cfg.UpdateNoteQueueName, err)
	}

	s.updateNoteQueue = updateNoteQueue

	deleteNoteQueue, err := ch.QueueDeclare(
		s.cfg.DeleteNoteQueueName, // name
		true,                      // durable
		false,                     // delete when unused
		false,                     // exclusive
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		return fmt.Errorf("error creating queue %s: %+v", s.cfg.DeleteNoteQueueName, err)
	}

	s.deleteNoteQueue = deleteNoteQueue

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
