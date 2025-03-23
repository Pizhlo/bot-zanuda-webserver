package worker

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type worker struct {
	conn    *amqp.Connection
	channel *amqp.Channel

	// queues
	createNoteQueue amqp.Queue
	updateNoteQueue amqp.Queue
}

const (
	createNoteQueueName = "create_note"
	updateNoteQueueName = "update_note"
)

func New(addr string) (*worker, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	createNoteQueue, err := ch.QueueDeclare(
		createNoteQueueName, // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("error creating queue %s: %+v", createNoteQueueName, err)
	}

	updateNoteQueue, err := ch.QueueDeclare(
		updateNoteQueueName, // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("error creating queue %s: %+v", updateNoteQueueName, err)
	}

	return &worker{
		conn:            conn,
		channel:         ch,
		createNoteQueue: createNoteQueue,
		updateNoteQueue: updateNoteQueue,
	}, nil
}

func (s *worker) Close() error {
	err := s.channel.Close()
	if err != nil {
		logrus.Errorf("worker: error closing channel rabbit mq: %+v", err)
	}

	return s.conn.Close()
}

func (s *worker) publish(queue string, body []byte) error {
	logrus.Debugf("rabbit: publishing message to queue '%s': %+v", queue, string(body))

	return s.channel.Publish(
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
