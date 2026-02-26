// Package cli implements the semtag command-line interface using stdlib flag.
package cli

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime/debug"

	ilog "github.com/flaticols/bump/internal/log"
	iterm "github.com/flaticols/bump/internal/term"
)

// Config holds all CLI configuration derived from flags and environment.
type Config struct {
	Repo    string
	Prefix  string
	Verbose bool
	Local   bool
	Brave   bool
	NoColor bool
	NoTTY   bool
	JSON    bool

	// Derived
	Interactive bool
	Term        iterm.Capability
	Color       bool
}

// Run is the main entry point for the CLI.
func Run() {
	cfg := parseFlags()

	if cfg.Repo != "" {
		if err := os.Chdir(path.Clean(cfg.Repo)); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
	}

	setupLogger(cfg)

	args := flag.Args()
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}

	if cfg.Brave {
		slog.Warn("brave mode enabled, ignoring warnings and errors")
	}
	if cfg.Verbose && cfg.Repo != "" {
		slog.Debug("working directory", "path", cfg.Repo)
	}

	var err error
	switch cmd {
	case "undo":
		err = runUndo(cfg)
	case "diff":
		err = runDiff(cfg, args[1:])
	case "version":
		printVersion()
		return
	case "help":
		printUsage()
		return
	default:
		err = runBump(cfg, args)
	}

	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Repo, "repo", "", "path to the repository")
	flag.StringVar(&cfg.Repo, "r", "", "path to the repository (shorthand)")
	flag.StringVar(&cfg.Prefix, "prefix", "", "tag prefix for multi-module repos (e.g., pkg/x)")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "enable verbose output")
	flag.BoolVar(&cfg.Local, "local", false, "skip remote operations")
	flag.BoolVar(&cfg.Local, "l", false, "skip remote operations (shorthand)")
	flag.BoolVar(&cfg.Brave, "brave", false, "skip all checks and confirmations")
	flag.BoolVar(&cfg.Brave, "b", false, "skip all checks and confirmations (shorthand)")
	flag.BoolVar(&cfg.NoColor, "no-color", false, "disable colored output")
	flag.BoolVar(&cfg.NoTTY, "no-tty", false, "disable interactive prompts")
	flag.BoolVar(&cfg.JSON, "json", false, "output JSON to stdout")

	flag.Usage = printUsage
	flag.Parse()

	cfg.Term = iterm.Detect(int(os.Stderr.Fd()))

	if os.Getenv("NO_COLOR") != "" {
		cfg.NoColor = true
	}
	cfg.Color = !cfg.NoColor && cfg.Term.IsTTY && cfg.Term.ColorLevel > 0
	cfg.Interactive = cfg.Term.IsTTY && !cfg.NoTTY && !cfg.Brave

	return cfg
}

func setupLogger(cfg *Config) {
	level := slog.LevelInfo
	if cfg.Verbose {
		level = slog.LevelDebug
	}

	styledHandler := ilog.NewStyledHandler(os.Stderr, level, !cfg.Color)

	if cfg.JSON {
		jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		slog.SetDefault(slog.New(slog.NewMultiHandler(styledHandler, jsonHandler)))
	} else {
		slog.SetDefault(slog.New(styledHandler))
	}
}

func printVersion() {
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" {
		fmt.Println(info.Main.Version)
	} else {
		fmt.Println("dev")
	}
}

func printUsage() {
	fmt.Fprint(os.Stderr, `semtag - Semantic versioning for Git

Usage:
  semtag [flags] [major|minor|patch] [package]
  semtag undo [flags]
  semtag diff [flags] <old-ref> [new-ref]

Commands:
  major        Bump major version (1.2.3 -> 2.0.0)
  minor        Bump minor version (1.2.3 -> 1.3.0)
  patch        Bump patch version (1.2.3 -> 1.2.4) [default]
  undo         Remove the latest semver tag
  diff         Compare Go API changes between refs
  version      Print version
  help         Show this help

Flags:
`)
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, `
Examples:
  semtag                           Bump patch version
  semtag major                     Bump major version
  semtag minor pkg/semver          Bump minor for pkg/semver module
  semtag --prefix pkg/x patch      Bump patch for pkg/x
  semtag --local                   Bump patch, skip remote
  semtag undo                      Remove latest tag
  semtag undo --brave              Remove latest tag without confirmation
  semtag diff v1.0.0 v1.1.0       Compare Go API between two refs
  semtag diff v1.0.0               Compare v1.0.0 against HEAD
  semtag --json                    Output structured JSON alongside styled stderr
`)
}
