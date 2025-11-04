package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/flaticols/bump/internal"
	"github.com/spf13/cobra"
)

func Run() {
	opts := createOptions()
	rootCmd := setupCommands(opts)

	color.NoColor = opts.NoColor
	printStartupMessages(opts)

	if err := internal.SetBumpWd(opts.RepoDirectory); err != nil {
		outputJSON(opts)
		opts.P.Println(opts.P.Err(err.Error()))
		os.Exit(1)
	}

	execErr := rootCmd.Execute()
	outputJSON(opts)

	if execErr != nil {
		opts.P.Println(opts.P.Err(execErr.Error()))
		os.Exit(1)
	}
}

// createOptions initializes and returns the Options struct with default values
func createOptions() *Options {
	return &Options{
		P: TextPrinters{
			Err:     color.New(color.FgRed).SprintfFunc(),
			Info:    color.New(color.FgBlue).SprintfFunc(),
			Warning: color.New(color.FgYellow).SprintfFunc(),
			Ok:      color.New(color.FgGreen).SprintfFunc(),
			Version: versionPrinter,
			Printf:  printfStderr,
			Println: printlnStderr,
			Symbols: Symbols{
				Ok:      color.New(color.FgGreen).Sprintf("•"),
				Warning: color.New(color.FgYellow).Sprintf("•"),
				Error:   color.New(color.FgRed).Sprintf("•"),
				Bullet:  color.New(color.FgWhite).Sprintf("•"),
			},
		},
		DefaultBranchs: []string{"latest", "main", "master", "develop"},
	}
}

// setupCommands configures the root command and its flags
func setupCommands(opts *Options) *cobra.Command {
	rootCmd := CreateRootCmd(opts)

	// Register persistent flags
	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&opts.RepoDirectory, "repo", "r", "", "path to the repository")
	pf.BoolVar(&opts.Verbose, "verbose", false, "enable verbose output")
	pf.BoolVarP(&opts.OnlyLocal, "local", "l", false, "if local is set, bump will not error if no remotes are found")
	pf.BoolVarP(&opts.BraveMode, "brave", "b", false, "if brave is set, bump will not ask any questions (default: false)")
	pf.BoolVar(&opts.NoColor, "no-color", false, "disable colorful output (default: false)")
	pf.BoolVar(&opts.JSON, "json", false, "output a single JSON object to stdout")
	pf.StringVar(&opts.Prefix, "prefix", "", "tag prefix to use (e.g., 'pkg/x')")

	rootCmd.ParseFlags(os.Args[1:])
	rootCmd.AddCommand(CreateUndoCmd(opts))

	return rootCmd
}

// printStartupMessages prints initial status messages if flags are enabled
func printStartupMessages(opts *Options) {
	if opts.BraveMode {
		opts.P.Printf("%s brave mode enabled, ignoring warnings and errors\n", opts.P.Symbols.Warning)
	}
	if opts.Verbose {
		opts.P.Printf("%s working directory: %s\n", opts.P.Symbols.Bullet, opts.RepoDirectory)
	}
}

// outputJSON outputs the result as JSON if JSON mode is enabled
func outputJSON(opts *Options) {
	if !opts.JSON {
		return
	}
	if opts.Result.Checks == nil {
		opts.Result.Checks = []string{}
	}
	b, _ := json.Marshal(opts.Result)
	fmt.Fprintln(os.Stdout, string(b))
}

// printfStderr writes formatted output to stderr
func printfStderr(format string, a ...any) {
	if _, err := fmt.Fprintf(os.Stderr, format, a...); err != nil {
		panic(err)
	}
}

// printlnStderr writes formatted output to stderr with a newline
func printlnStderr(format string, a ...any) {
	if _, err := fmt.Fprintln(os.Stderr, fmt.Sprintf(format, a...)); err != nil {
		panic(err)
	}
}

// versionPrinter formats a version string with "v" prefix
func versionPrinter(ver string) string {
	return "v" + ver
}
