package chat

type Client interface {
	Send(message string) error
	Upload(path string) error
	Close() error
	Events() <-chan Event
	Done() <-chan struct{}
}
