//go:generate stringer -type=ConnectionStatus

package types

type ConnectionStatus int64

const (
	NOT_CONNECTED ConnectionStatus = iota
	CONNECTED
	LOST
)

