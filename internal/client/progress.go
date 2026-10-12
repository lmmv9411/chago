package client

import (
	"io"
	"time"

	"github.com/lmmv9411/chago/internal/events"
)

type ProgressWriter struct {
	total          int64
	writer         io.Writer
	written        int64
	id             string
	filename       string
	events         chan<- events.Event
	hasReported    bool
	lastProgressAt time.Time
}

func (p *ProgressWriter) Write(data []byte) (int, error) {
	n, err := p.writer.Write(data)
	if err != nil {
		return n, err
	}

	p.written += int64(n)
	if p.total > 0 {

		now := time.Now()
		ratio := float64(p.written) / float64(p.total)
		completed := p.written >= p.total

		if !p.hasReported || now.Sub(p.lastProgressAt) >= 150*time.Millisecond || completed {

			p.events <- events.Event{
				Kind:       events.TransferProgress,
				TransferId: p.id,
				FileName:   p.filename,
				Ratio:      ratio,
			}
			p.lastProgressAt = now
			p.hasReported = true
		}
	}

	return n, nil
}
