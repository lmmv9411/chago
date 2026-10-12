package events

type Kind uint8

const (
	ChatMessage = iota
	Error
	TransferProgress
	TransferDone
)

type Event struct {
	Kind       Kind
	Message    string
	TransferId string
	FileName   string
	Ratio      float64
}
