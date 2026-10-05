package serverchat

import "net"

type EventType int

const (
	connection EventType = iota
	disconnection
	message
)

type Event struct {
	kind    EventType
	address string
	conn    net.Conn
	message string
}
