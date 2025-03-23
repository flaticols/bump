package main

import (
	"github.com/fatih/color"
	"github.com/flaticols/bump/cmd"
	"github.com/flaticols/bump/internal"
	"os"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	opts := &cmd.Options{
		ErrPrinter:     color.New(color.FgRed).SprintfFunc(),
		InfoPrinter:    color.New(color.FgWhite).SprintfFunc(),
		WarningPrinter: color.New(color.FgYellow).SprintfFunc(),
		OkPrinter:      color.New(color.FgGreen).SprintfFunc(),
		GitDetailer:    &internal.GitState{},
	}

	rootCmd := cmd.CreateRootCmd(opts)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	//color.Green("Successfully bumped version to v%s", ver.String())
}
