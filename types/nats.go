package types

import "github.com/nats-io/stan.go"

type NotConnectedError struct {
}

func (e *NotConnectedError) Error() string {
	return "not connected"
}

type ConnectionEstablishedError struct {
}

func (e *ConnectionEstablishedError) Error() string {
	return "connected established"
}

type IConnectionManager interface {
	Open() (err error)
	Close() (err error)
	GetConn() (stan.Conn, error)
	SubscribeToStatusChanges() <-chan ConnectionStatus
}

type ISubscriptionManager interface {
	Subscribe() error
	Close() error
}
