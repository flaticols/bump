// Package log provides a custom slog handler for styled terminal output
// and a MultiHandler for fan-out to multiple slog handlers (e.g. styled + JSON).
package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	iterm "github.com/flaticols/bump/internal/term"
)

// StyledHandler outputs slog records as styled terminal lines with colored bullets.
// Output format: ● message [key=value ...]
type StyledHandler struct {
	w       io.Writer
	level   slog.Level
	noColor bool
	mu      sync.Mutex
	attrs   []slog.Attr
	groups  []string
}

// NewStyledHandler creates a handler that writes styled output to w.
func NewStyledHandler(w io.Writer, level slog.Level, noColor bool) *StyledHandler {
	return &StyledHandler{w: w, level: level, noColor: noColor}
}

func (h *StyledHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *StyledHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	bullet := h.bullet(r.Level)
	var b strings.Builder
	b.WriteString(bullet)
	b.WriteByte(' ')
	b.WriteString(r.Message)

	// Append pre-set attrs
	for _, a := range h.attrs {
		h.writeAttr(&b, a)
	}

	// Append record attrs
	r.Attrs(func(a slog.Attr) bool {
		h.writeAttr(&b, a)
		return true
	})

	b.WriteByte('\n')
	_, err := io.WriteString(h.w, b.String())
	return err
}

func (h *StyledHandler) writeAttr(b *strings.Builder, a slog.Attr) {
	if a.Equal(slog.Attr{}) {
		return
	}
	val := a.Value.Resolve().String()
	colored := iterm.Colorize(iterm.Cyan, val, !h.noColor)
	fmt.Fprintf(b, " %s=%s", a.Key, colored)
}

func (h *StyledHandler) bullet(level slog.Level) string {
	color := iterm.Green
	switch {
	case level >= slog.LevelError:
		color = iterm.Red
	case level >= slog.LevelWarn:
		color = iterm.Yellow
	case level < slog.LevelInfo:
		color = iterm.Gray
	}
	return iterm.Colorize(color, "●", !h.noColor)
}

func (h *StyledHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &StyledHandler{
		w: h.w, level: h.level, noColor: h.noColor,
		attrs:  append(append([]slog.Attr{}, h.attrs...), attrs...),
		groups: h.groups,
	}
}

func (h *StyledHandler) WithGroup(name string) slog.Handler {
	return &StyledHandler{
		w: h.w, level: h.level, noColor: h.noColor,
		attrs:  h.attrs,
		groups: append(append([]string{}, h.groups...), name),
	}
}

// MultiHandler fans out log records to multiple handlers.
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler creates a handler that writes to all given handlers.
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: handlers}
}
