package main

import (
	"context"
	"github.com/mvcatsifma/nats-streaming-client/nats"
	"github.com/mvcatsifma/nats-streaming-client/types"
	"github.com/nats-io/stan.go"
	"log"
	"os"
	"os/signal"
)

const clientId = "test-sub"

func main() {
	conn := nats.NewConnectionManager(types.ClusterID, clientId)

	ctx, cancel := context.WithCancel(context.Background())
	go connectionStatusWorker(ctx, conn)

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

	cancel()

	err := conn.Conn.Close()
	if err != nil {
		log.Println(err)
	}
}

func connectionStatusWorker(parent context.Context, conn types.IConnectionManager) {
	var ctx context.Context
	var cancel context.CancelFunc
	sub := conn.SubscribeToStatusChanges()
	for {
		status := <-sub
		switch status {
		case types.NOT_CONNECTED:
			log.Println("not connected")
			ctx, cancel = context.WithCancel(parent)
			go nats.OpenWithRetry(ctx, conn)
		case types.CONNECTED:
			log.Println("connected")
			cancel()
			conn, err := conn.GetConn()
			if err != nil {
				log.Fatal(err)
			}
			initSubscriptions(conn)
		case types.LOST:
			log.Println("lost")
			ctx, cancel = context.WithCancel(parent)
			go nats.OpenWithRetry(ctx, conn)
		default:
			log.Printf("unknown status: %v", status)
		}
	}
}

func initSubscriptions(conn stan.Conn) {
	var subs []types.ISubscriptionManager
	subs = append(subs, nats.NewDefaultHandler(conn, "main", "test-channel"))

	for _, sub := range subs {
		err := sub.Subscribe()
		if err != nil {
			log.Printf("error subscribing: %v", err)
		}
	}
}
