package tui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
)

// Select prompts the user to pick one item from options.
// Falls back to a plain numbered list when not interactive.
func Select(title string, options []string, interactive bool) (string, error) {
	if !interactive {
		return selectPlain(os.Stdin, os.Stderr, title, options)
	}

	huhOpts := make([]huh.Option[string], len(options))
	for i, o := range options {
		huhOpts[i] = huh.NewOption(o, o)
	}

	var selected string
	err := huh.NewSelect[string]().
		Title(title).
		Options(huhOpts...).
		Value(&selected).
		WithTheme(huh.ThemeBase()).Run()
	if err != nil {
		return "", err
	}
	return selected, nil
}

func selectPlain(r io.Reader, w io.Writer, title string, options []string) (string, error) {
	fmt.Fprintln(w, title)
	for i, o := range options {
		fmt.Fprintf(w, "  %d) %s\n", i+1, o)
	}
	fmt.Fprintf(w, "Choice [1-%d]: ", len(options))

	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || n < 1 || n > len(options) {
			return "", fmt.Errorf("invalid selection: %s", scanner.Text())
		}
		return options[n-1], nil
	}
	return "", fmt.Errorf("no input")
}
