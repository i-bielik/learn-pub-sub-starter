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

	// publish channel
	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal("Failed to parse username:", err)
	}

	// new game state
	gs := gamelogic.NewGameState(username)

	// subscribe to direct exchange
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gs.GetUsername(),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	// subscribe to army moves
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gs.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerMove(gs, publishCh),
	)
	if err != nil {
		log.Fatalf("could not subscribe to army moves: %v", err)
	}

	// subscribe to war topic
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		string(routing.WarRecognitionsPrefix)+".*",
		pubsub.Durable,
		handlerWar(gs),
	)
	if err != nil {
		log.Fatalf("could not subscribe to war declarations: %v", err)
	}

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
			// record a move
			mv, err := gs.CommandMove(cmd)
			if err != nil {
				fmt.Println(err)
				continue
			}
			// publish the move to peril_topic
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+gs.GetUsername(),
				mv,
			)
			if err != nil {
				fmt.Println("Failed to publish a move: %w", err)
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
