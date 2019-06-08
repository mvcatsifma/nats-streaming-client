package main

import (
	"fmt"
	"github.com/mvcatsifma/nats-streaming-client/types"
	"github.com/nats-io/stan.go"
	"log"
	"os"
	"os/signal"
)

const clientId = "test-sub"

func main() {
	sc, err := stan.Connect(types.ClusterID, clientId)
	if err != nil {
		log.Fatal(err)
	}
	_, err = sc.Subscribe("test-channel", func(msg *stan.Msg) {
		log.Printf("%#v\n", msg)
	}, stan.DurableName("main"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("subscribed to channel test-channel")

	osInterruptChannel := make(chan os.Signal, 1)
	signal.Notify(osInterruptChannel, os.Interrupt, os.Kill)

	// This read will block execution until an OS signal (such as Ctrl-C) is received:
readInterruptLoop:
	for {
		select {
		case sig := <-osInterruptChannel:
			switch sig {
			case os.Kill: // KILL received
				log.Println("SIGINT received, shutting down")
				fallthrough
			case os.Interrupt: // SIGINT received
				log.Println("SIGINT received, shutting down")
				break readInterruptLoop
			}
		}
	}
	signal.Stop(osInterruptChannel) // stop sending signals

	err = sc.Close()
	if err != nil {
		log.Println(err)
	}
}



