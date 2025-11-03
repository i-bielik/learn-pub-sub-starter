package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
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
			switch handler(msg) {
			case Ack:
				// ... this consumer is responsible for sending message per log
				err = message.Ack(false)
				if err != nil {
					log.Printf("ack error: %+v", err)
				}
			case NackDiscard:
				err = message.Nack(false, false)
				if err != nil {
					log.Printf("nack error: %+v", err)
				}
			case NackRequeue:
				err = message.Nack(false, true)
				if err != nil {
					log.Printf("nack error: %+v", err)
				}
			}

		}
	}()
	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
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
			// unmarshall to T message from gob
			var msg T
			b := bytes.NewBuffer(message.Body)
			dec := gob.NewDecoder(b)
			err := dec.Decode(&msg)
			if err != nil {
				log.Printf("unmarshalling error: %v", err)
				continue
			}
			// call handler func
			switch handler(msg) {
			case Ack:
				// ... this consumer is responsible for sending message per log
				err = message.Ack(false)
				if err != nil {
					log.Printf("ack error: %+v", err)
				}
			case NackDiscard:
				err = message.Nack(false, false)
				if err != nil {
					log.Printf("nack error: %+v", err)
				}
			case NackRequeue:
				err = message.Nack(false, true)
				if err != nil {
					log.Printf("nack error: %+v", err)
				}
			}

		}
	}()
	return nil
}
