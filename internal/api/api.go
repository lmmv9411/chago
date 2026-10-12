package api

import (
	"github.com/lmmv9411/chago/internal/events"
)

type Api interface {
	Send(message string) error
	Upload(path string) error
	Close() error
	Events() chan events.Event
}
