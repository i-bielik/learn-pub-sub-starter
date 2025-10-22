package pubsub

import (
	"context"
	"encoding/json"
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
