// Package term provides terminal capability detection and ANSI color utilities.
// It uses golang.org/x/term for TTY detection and respects NO_COLOR, TERM=dumb,
// and COLORTERM environment variables.
package term

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// ColorLevel indicates the terminal's color support.
const (
	ColorNone    = 0 // No color support
	ColorBasic   = 1 // Basic 16 colors
	Color256     = 2 // 256 colors
	ColorTrue    = 3 // 24-bit true color
)

// Capability holds detected terminal capabilities.
type Capability struct {
	IsTTY      bool
	ColorLevel int
	Width      int
	Height     int
}

// Detect probes the given file descriptor for terminal capabilities.
func Detect(fd int) Capability {
	c := Capability{}
	c.IsTTY = term.IsTerminal(fd)
	if !c.IsTTY {
		return c
	}
	if w, h, err := term.GetSize(fd); err == nil {
		c.Width = w
		c.Height = h
	}
	c.ColorLevel = detectColorLevel()
	return c
}

func detectColorLevel() int {
	if os.Getenv("NO_COLOR") != "" {
		return ColorNone
	}
	if os.Getenv("TERM") == "dumb" {
		return ColorNone
	}
	ct := os.Getenv("COLORTERM")
	if ct == "truecolor" || ct == "24bit" {
		return ColorTrue
	}
	t := os.Getenv("TERM")
	if strings.Contains(t, "256color") {
		return Color256
	}
	return ColorBasic
}

// ANSI escape codes for basic colors.
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Gray    = "\033[90m"
)

// Colorize wraps text with the given ANSI color code. Returns plain text when disabled.
func Colorize(color, text string, enabled bool) string {
	if !enabled {
		return text
	}
	return color + text + Reset
}
