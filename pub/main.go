package main

import (
	"github.com/mvcatsifma/nats-streaming-client/types"
	"github.com/nats-io/stan.go"
	"log"
)

const clientId = "test-pub"

func main() {
	sc, err := stan.Connect(types.ClusterID, clientId)
	if err != nil {
		log.Fatal(err)
	}

	err = sc.Publish("test-channel", []byte("hello world"))
	if err != nil {
		log.Fatal(err)
	}
}
