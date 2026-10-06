package client

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Terminal struct {
	mu           sync.Mutex
	progress     *ProgressState
	progressDone chan struct{}
}

type ProgressState struct {
	active   bool
	total    int64
	written  int64
	barWidth int
	mu       sync.RWMutex
}

func (t *Terminal) Print(format string, args ...any) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.progress != nil {
		fmt.Print("\r\033[2K")
	}

	fmt.Printf(format, args...)

	if t.progress != nil {
		fmt.Printf("\r%s", t.progress.String())
	}
}

func (t *Terminal) PrintProgress() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.progress == nil {
		return
	}

	fmt.Printf("\r\033[2K%s", t.progress.String())
}

func (t *Terminal) SetProgress(state *ProgressState) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state.active = true
	t.progress = state
}

func (t *Terminal) StartProgress() {
	t.mu.Lock()

	if t.progressDone != nil {
		t.mu.Unlock()
		return
	}

	done := make(chan struct{})
	t.progressDone = done

	t.mu.Unlock()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				t.PrintProgress()

			case <-done:
				return
			}
		}
	}()
}

func (t *Terminal) StopProgress() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.progressDone == nil {
		return
	}

	close(t.progressDone)
	t.progressDone = nil
	t.progress = nil
	t.progress.active = false
}

func (t *Terminal) ClearProgressLine() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.progress == nil {
		return
	}

	fmt.Print("\r\033[2K")
}

func (p *ProgressState) String() string {

	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.active || p.total == 0 {
		return ""
	}

	percentage := float64(p.written) / float64(p.total) * 100
	filled := int((percentage / 100.0) * float64(p.barWidth))

	var bar strings.Builder

	bar.WriteString("[")

	for i := range p.barWidth {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}

	fmt.Fprintf(&bar, "] %.0f%%", percentage)

	return bar.String()
}
