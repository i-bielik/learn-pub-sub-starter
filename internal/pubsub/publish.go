package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	// Marshal value to JSON bytes
	body, err := json.Marshal(val)
	if err != nil {
		return err
	}

	// Prepare message
	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	if err != nil {
		log.Fatalf("basic.publish: %v", err)
		return err
	}

	return nil

}

// Define an enum-like type for queue types as strings
type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	// create new channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to create new channel:", err)
	}

	// Declare the queue based on type
	var q amqp.Queue
	if queueType == Durable {
		q, err = ch.QueueDeclare(
			queueName,
			true,  // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,   // args
		)
	} else {
		q, err = ch.QueueDeclare(
			queueName,
			false, // durable
			true,  // autoDelete
			true,  // exclusive
			false, // noWait
			nil,   // args
		)
	}
	if err != nil {
		log.Fatal("Failed to declare queue:", err)
		return nil, amqp.Queue{}, err
	}

	// bind queue to channel
	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		log.Fatal("Failed to bind queue to channel:", err)
		return nil, amqp.Queue{}, err
	}

	return ch, q, nil

}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("created error: %w", err)
	}
	channel, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume error: %w", err)
	}
	go func() {
		for message := range channel {
			// unmarshall to T message
			var msg T
			err := json.Unmarshal(message.Body, &msg)
			if err != nil {
				log.Printf("unmarshalling error: %v", err)
				continue
			}
			// call handler func
			handler(msg)
			// ... this consumer is responsible for sending message per log
			if e := message.Ack(false); e != nil {
				log.Printf("ack error: %+v", e)
			}
		}
	}()
	return nil
}
