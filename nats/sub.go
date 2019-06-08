package nats

import (
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
	"log"
)

type sub struct {
	Channel     string              // name of the nats channel to subscribe too
	Conn        stan.Conn           // the connection to the nats server
	DurableName string              // this subscriptions durable name
	Handler     func(msg *stan.Msg) // function that will handle messages for this subscription
	sub         stan.Subscription   // the managed subscription
}

func (s *sub) Subscribe() error {
	sub, err := s.Conn.Subscribe(s.Channel, s.Handler, stan.DurableName(s.DurableName), stan.StartAt(pb.StartPosition_NewOnly))
	if err != nil {
		return err
	}
	s.sub = sub
	return nil
}

func (s *sub) Close() error {
	err := s.sub.Close()
	if err != nil {
		return err
	}
	return nil
}

func NewSubscriptionManager(conn stan.Conn, name string, channel string, handler func(msg *stan.Msg)) *sub {
	return &sub{
		Channel:     channel,
		Conn:        conn,
		DurableName: name,
		Handler:     handler,
	}
}

var defaultHandler = func(msg *stan.Msg) {
	log.Printf("%#v\n", msg)
}

func NewDefaultHandler(conn stan.Conn, name string, channel string) *sub {
	return NewSubscriptionManager(conn, name, channel, defaultHandler)
}
