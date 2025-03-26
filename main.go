package main

import (
	"fmt"
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
		P: cmd.TextPrinters{
			Err:     color.New(color.FgRed).SprintfFunc(),
			Info:    color.New(color.FgBlue).SprintfFunc(),
			Warning: color.New(color.FgYellow).SprintfFunc(),
			Ok:      color.New(color.FgGreen).SprintfFunc(),
			Version: versionPrinter,
			Symbols: cmd.Symbols{
				Ok:      color.New(color.FgGreen).Sprintf("✓"),
				Warning: color.New(color.FgYellow).Sprintf("⚠"),
				Error:   color.New(color.FgRed).Sprintf("✗"),
				Bullet:  color.New(color.FgWhite).Sprintf("•"),
			},
		},
		GitDetailer: &internal.GitState{},
	}

	// Create the root command
	rootCmd := cmd.CreateRootCmd(opts)

	rootCmd.PersistentFlags().StringVarP(&opts.RepoDirectory, "repo", "r", "", "path to the repository")
	rootCmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&opts.LocalRepo, "local", "l", false, "if local is set, bump will not error if no remotes are found")
	rootCmd.PersistentFlags().BoolVarP(&opts.BraveMode, "brave", "b", false, "if brave is set, bump will not ask any questions (default: false)")
	rootCmd.PersistentFlags().BoolVar(&opts.NoColor, "no-color", false, "disable colorful output (default: false)")

	undoCmd := cmd.CreateUndoCmd(opts)
	rootCmd.AddCommand(undoCmd)

	color.NoColor = opts.NoColor

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func versionPrinter(ver string) string {
	return fmt.Sprintf("v%s", ver)
}
