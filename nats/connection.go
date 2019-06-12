package nats

import (
	"context"
	"github.com/mvcatsifma/nats-streaming-client/types"
	"github.com/nats-io/stan.go"
	"log"
	"time"
)

type manager struct {
	ClientId                string                 // the client id to use while connecting
	ClusterId               string                 // the id of the nats cluster to connect to
	Conn                    stan.Conn              // the managed connection
	Status                  types.ConnectionStatus // the status of the connection
	statusChangeChan        chan<- types.ConnectionStatus
	statusChangeSubscribers map[chan types.ConnectionStatus]bool
}

func NewConnectionManager(clusterId string, clientId string) *manager {
	initialStatus := types.NOT_CONNECTED
	result := &manager{
		ClusterId:               clusterId,
		ClientId:                clientId,
		Status:                  initialStatus,
		statusChangeChan:        make(chan types.ConnectionStatus, 1),
		statusChangeSubscribers: make(map[chan types.ConnectionStatus]bool),
	}
	result.sendStatusChange(initialStatus) // send initial status
	return result
}

func (m *manager) Open() (err error) {
	if m.Status == types.CONNECTED {
		return nil
	}
	lostHandler := stan.SetConnectionLostHandler(func(conn stan.Conn, e error) {
		newStatus := types.LOST
		m.Status = newStatus
		m.sendStatusChange(newStatus)
	})
	pings := stan.Pings(1, 3)

	conn, err := stan.Connect(m.ClusterId, m.ClientId, pings, lostHandler)
	if err != nil {
		return err
	}

	m.Conn = conn
	newStatus := types.CONNECTED
	m.Status = newStatus
	m.sendStatusChange(newStatus)

	return nil
}

func (m *manager) Close() (err error) {
	if m.Conn == nil {
		return &types.NotConnectedError{}
	}
	err = m.Conn.Close()
	if err != nil {
		return err
	}
	newStatus := types.NOT_CONNECTED
	m.Status = newStatus
	m.sendStatusChange(newStatus)
	m.Conn = nil

	for sub := range m.statusChangeSubscribers {
		close(sub)
		delete(m.statusChangeSubscribers, sub)
	}

	return nil
}

func (m *manager) SubscribeToStatusChanges() <-chan types.ConnectionStatus {
	sub := make(chan types.ConnectionStatus, 1)
	m.statusChangeSubscribers[sub] = true
	sub <- m.Status // send current status
	return sub
}

func (m *manager) GetConn() (stan.Conn, error) {
	if m.Conn == nil {
		return nil, &types.NotConnectedError{}
	}
	return m.Conn, nil
}

func (m *manager) sendStatusChange(newStatus types.ConnectionStatus) {
	for sub := range m.statusChangeSubscribers {
		select {
		case sub <- newStatus:
			log.Printf("write status %v\n", newStatus.String())
		case <-time.Tick(1 * time.Second): // close and remove slow subscribers
			close(sub)
			delete(m.statusChangeSubscribers, sub)
		}
	}
}

func OpenWithRetry(ctx context.Context, conn types.IConnectionManager) {
	err := conn.Open()
	if err == nil {
		return
	}
	ticker := time.NewTicker(5 * time.Second)
ConnectLoop:
	for {
		select {
		case <-ticker.C:
			err := conn.Open()
			if err != nil {
				log.Printf("connect error: %v", err)
				continue
			} else {
				break ConnectLoop
			}
		case <-ctx.Done():
			log.Println("cancel")
			break ConnectLoop
		}
	}
	ticker.Stop()
}
