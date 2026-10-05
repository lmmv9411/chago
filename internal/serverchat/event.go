package serverchat

type EventType int

const (
	connection EventType = iota
	disconnection
	message
	notification
)

type Event struct {
	kind    EventType
	address string
	conn    *Connection
	message string
	status  Status
}
