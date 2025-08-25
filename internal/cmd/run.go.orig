package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/flaticols/bump/internal"
)

func Run() {
	// Create color printers for formatted output
	opts := &Options{
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

	rootCmd := CreateRootCmd(opts)

	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&opts.RepoDirectory, "repo", "r", "", "path to the repository")
	pf.BoolVar(&opts.Verbose, "verbose", false, "enable verbose output")
	pf.BoolVarP(&opts.OnlyLocal, "local", "l", false, "if local is set, bump will not error if no remotes are found")
	pf.BoolVarP(&opts.BraveMode, "brave", "b", false, "if brave is set, bump will not ask any questions (default: false)")
	pf.BoolVar(&opts.NoColor, "no-color", false, "disable colorful output (default: false)")
	pf.StringVar(&opts.Prefix, "prefix", "", "tag prefix to use (e.g., 'pkg/x')")
	rootCmd.ParseFlags(os.Args[1:])

	opts.Exit = func() {
		if !opts.BraveMode {
			os.Exit(1)
		}

		os.Exit(0)
	}

	undoCmd := CreateUndoCmd(opts)
	rootCmd.AddCommand(undoCmd)

	color.NoColor = opts.NoColor

	if opts.BraveMode {
		opts.P.Printf("%s brave mode enabled, ignoring warnings and errors\n", opts.P.Symbols.Warning)
	}

	if opts.Verbose {
		opts.P.Printf("%s working directory: %s\n", opts.P.Symbols.Bullet, opts.RepoDirectory)
	}

	err := internal.SetBumpWd(opts.RepoDirectory)
	if err != nil {
		opts.P.Println(opts.P.Err(err.Error()))
		os.Exit(1)
	}

	rootCmd.ErrOrStderr()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// printfStderr writes a formatted string to the standard error output (os.Stderr).
// It takes a format string and a variadic number of arguments to format the output.
// If an error occurs during writing, the function panics with the encountered error.
//
// Parameters:
//   - format: A string specifying the format of the output, similar to fmt.Sprintf.
//   - a: Variadic arguments to be formatted according to the format string.
func printfStderr(format string, a ...any) {
	_, err := fmt.Fprintf(os.Stderr, format, a...)
	if err != nil {
		panic(err)
	}
}

// printlnStderr writes a formatted string to the standard error output (os.Stderr).
// It takes a format string and a variadic list of arguments, similar to fmt.Sprintf.
// If an error occurs while writing to os.Stderr, the function will panic.
//
// Parameters:
//   - format: A string containing the text to be formatted.
//   - a: A variadic list of arguments to be formatted into the string.
func printlnStderr(format string, a ...any) {
	_, err := fmt.Fprintln(os.Stderr, fmt.Sprintf(format, a...))
	if err != nil {
		panic(err)
	}
}

// versionPrinter formats the given version string by prefixing it with "v".
//
// Parameters:
//   - ver: A string representing the version number.
//
// Returns:
//
//	A formatted string with the version number prefixed by "v".
func versionPrinter(ver string) string {
	return fmt.Sprintf("v%s", ver)
}
