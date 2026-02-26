package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/flaticols/bump/internal/apidiff"
)

func runDiff(cfg *Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: semtag diff <old-ref> [new-ref]\n  old-ref: git tag, branch, or commit\n  new-ref: git tag, branch, or commit (default: HEAD)")
	}

	oldRef := args[0]
	newRef := "HEAD"
	if len(args) >= 2 {
		newRef = args[1]
	}

	slog.Info(fmt.Sprintf("comparing API: %s -> %s", oldRef, newRef))
	slog.Debug("creating temporary worktrees (working tree untouched)")

	report, err := apidiff.Compare(oldRef, newRef)
	if err != nil {
		return fmt.Errorf("api diff: %w", err)
	}

	report.WriteText(os.Stdout, true)

	slog.Info(fmt.Sprintf("summary: %s", report.Summary()))
	slog.Info(fmt.Sprintf("suggested bump: %s", report.SuggestedBump()))

	if report.HasBreaking() {
		os.Exit(1)
	}

	return nil
}
