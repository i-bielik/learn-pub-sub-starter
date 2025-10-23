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
	fmt.Println("Starting Peril client...")
	conn_string := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(conn_string)

	if err != nil {
		fmt.Println("Failed to connect to AMPQ:", err)
		os.Exit(1)
	}

	defer conn.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal("Failed to parse username:", err)
	}

	_, _, err = pubsub.DeclareAndBind(
		conn,
		string(routing.ExchangePerilDirect),
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Fatal("Failed to create channel and queue:", err)
	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	// if signal received, print message and close connection
	fmt.Println("Received an interrupt, closing connections...")
}
