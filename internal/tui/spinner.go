package tui

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Spinner shows a simple animated spinner for long-running operations.
// Falls back to a static message when not interactive.
type Spinner struct {
	w           io.Writer
	msg         string
	done        chan struct{}
	mu          sync.Mutex
	interactive bool
}

// NewSpinner creates a spinner that writes to w.
func NewSpinner(w io.Writer, msg string, interactive bool) *Spinner {
	return &Spinner{w: w, msg: msg, done: make(chan struct{}), interactive: interactive}
}

// Start begins the spinner animation in a goroutine.
func (s *Spinner) Start() {
	if !s.interactive {
		fmt.Fprintf(s.w, "%s...\n", s.msg)
		return
	}
	go s.run()
}

// Stop ends the spinner and clears the line.
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-s.done:
	default:
		close(s.done)
	}
	if s.interactive {
		fmt.Fprintf(s.w, "\r\033[K")
	}
}

func (s *Spinner) run() {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			fmt.Fprintf(s.w, "\r%s %s", frames[i%len(frames)], s.msg)
			i++
		}
	}
}
