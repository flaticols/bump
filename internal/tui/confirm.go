// Package tui provides interactive terminal UI elements backed by charmbracelet.
// When the terminal is not interactive (no TTY or --no-tty), it falls back to
// plain stdin/stdout prompts.
package tui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
)

// ConfirmOpt configures a confirmation prompt.
type ConfirmOpt func(*confirmOpts)

type confirmOpts struct {
	yesText     string
	noText      string
	bypass      bool
	bypassVal   bool
	interactive bool
}

// WithYes sets the affirmative button text.
func WithYes(text string) ConfirmOpt {
	return func(o *confirmOpts) { o.yesText = text }
}

// WithNo sets the negative button text.
func WithNo(text string) ConfirmOpt {
	return func(o *confirmOpts) { o.noText = text }
}

// BypassIf skips the prompt and returns val when cond is true.
func BypassIf(cond, val bool) ConfirmOpt {
	return func(o *confirmOpts) { o.bypass = cond; o.bypassVal = val }
}

// Interactive controls whether to use the full TUI or plain text prompt.
func Interactive(v bool) ConfirmOpt {
	return func(o *confirmOpts) { o.interactive = v }
}

// Confirm asks the user a yes/no question and returns the answer.
func Confirm(question string, opts ...ConfirmOpt) bool {
	o := confirmOpts{
		yesText:     "Yes",
		noText:      "No",
		interactive: true,
	}
	for _, opt := range opts {
		opt(&o)
	}

	if o.bypass {
		return o.bypassVal
	}

	if !o.interactive {
		return confirmPlain(os.Stdin, os.Stderr, question)
	}

	var confirm bool
	err := huh.NewConfirm().
		Title(question).
		Affirmative(o.yesText).
		Negative(o.noText).
		Value(&confirm).
		WithTheme(huh.ThemeBase()).Run()
	if err != nil {
		return false
	}
	return confirm
}

func confirmPlain(r io.Reader, w io.Writer, question string) bool {
	fmt.Fprintf(w, "%s [y/N]: ", question)
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		ans := strings.ToLower(strings.TrimSpace(scanner.Text()))
		return ans == "y" || ans == "yes"
	}
	return false
}
