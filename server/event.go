package server

import "net"

type EventType int

const (
	connection EventType = iota
	disconnection
	message
	file
)

type FileEvent struct {
	name string
	size int64
	path string
}

type Event struct {
	kind    EventType
	address string
	conn    net.Conn
	message string
	user    string
	file    FileEvent
}
