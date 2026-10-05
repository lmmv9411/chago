package serverchat

type EventType int

const (
	connection EventType = iota
	disconnection
	message
)

type Event struct {
	kind    EventType
	address string
	conn    *Connection
	message string
}
