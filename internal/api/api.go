package api

type Api interface {
	Send(message string) error
	Upload(path string) error
	Close() error
	Events() <-chan Event
}
