package main

import (
	"os"

	"github.com/fatih/color"
	"github.com/flaticols/bump/cmd"
	"github.com/flaticols/bump/internal"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	// Create color printers for formatted output
	opts := &cmd.Options{
		ErrPrinter:     color.New(color.FgRed).SprintfFunc(),
		InfoPrinter:    color.New(color.FgBlue).SprintfFunc(),
		WarningPrinter: color.New(color.FgYellow).SprintfFunc(),
		OkPrinter:      color.New(color.FgGreen).SprintfFunc(),
		GitDetailer:    &internal.GitState{},
	}

	// Create the root command
	rootCmd := cmd.CreateRootCmd(opts)
	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
