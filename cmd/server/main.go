package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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

	// print server help at the beginning
	gamelogic.PrintServerHelp()

	// Declare and bind new queue
	// ch, _, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, "game_logs", routing.GameLogSlug, pubsub.Durable)

	// create new channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to create new channel:", err)
	}

	fmt.Println("AMPQ connection successful")

out:
	for {
		cmd := gamelogic.GetInput()
		switch cmd[0] {
		case "":
			continue
		case "pause":
			log.Println("Sending a Pause message")

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
		case "resume":
			log.Println("Sending a Resume message")
			// publish message
			err = pubsub.PublishJSON(
				ch,
				string(routing.ExchangePerilDirect),
				string(routing.PauseKey),
				routing.PlayingState{IsPaused: false},
			)
			if err != nil {
				log.Fatal("Failed to publish message:", err)
			}
		case "quit":
			log.Println("Exiting")
			break out // break out of the loop
		default:
			log.Println("Not recognized command. Please retry.")

		}

	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	// if signal received, print message and close connection
	fmt.Println("Received an interrupt, closing connections...")

}
