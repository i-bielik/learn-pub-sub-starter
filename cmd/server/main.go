package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	conn_string := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(conn_string)

	if err != nil {
		fmt.Println("Failed to connect to AMPQ:", err)
		os.Exit(1)
	}

	defer conn.Close()

	// create new channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to create new channel:", err)
	}

	fmt.Println("AMPQ connection successful")

	// publish message
	err = pubsub.PublishJSON(
		ch,
		string(routing.ExchangePerilDirect),
		string(routing.PauseKey),
		routing.PlayingState{IsPaused: true},
	)
	if err != nil {
		log.Fatal("Failed to publish message:", err)
	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	// if signal received, print message and close connection
	fmt.Println("Received an interrupt, closing connections...")

}
