package main

import (
	"fmt"
	"log"
	"os"

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

	// new game state
	gs := gamelogic.NewGameState(username)

	// subscribe
	pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gs.GetUsername(),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gs),
	)

	for {
		cmd := gamelogic.GetInput()
		switch cmd[0] {
		case "":
			continue
		case "spawn":
			err = gs.CommandSpawn(cmd)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "move":
			_, err := gs.CommandMove(cmd)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Move successful")
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Not recognized command. Please retry.")
		}

	}
}
