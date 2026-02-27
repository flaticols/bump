package git

import (
	"fmt"
	"os"
	"os/exec"
)

// Worktree represents a temporary git worktree used for operations
// that must not touch the current working tree (e.g., API diff).
type Worktree struct {
	Path string
	Ref  string
}

// CreateTempWorktree creates a detached git worktree in a temp directory
// checked out at the given ref (tag, branch, or commit).
func CreateTempWorktree(ref string) (*Worktree, error) {
	dir, err := os.MkdirTemp("", "semtag-wt-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	cmd := exec.Command("git", "worktree", "add", "--detach", dir, ref)
	if out, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(dir)
		return nil, fmt.Errorf("create worktree at %s: %s: %w", ref, string(out), err)
	}

	return &Worktree{Path: dir, Ref: ref}, nil
}

// Remove cleans up the worktree directory and its git metadata.
func (w *Worktree) Remove() error {
	cmd := exec.Command("git", "worktree", "remove", "--force", w.Path)
	if out, err := cmd.CombinedOutput(); err != nil {
		// Fallback: manual cleanup if the command fails (e.g. older git).
		os.RemoveAll(w.Path)
		exec.Command("git", "worktree", "prune").Run() //nolint:errcheck
		return fmt.Errorf("remove worktree %s: %s: %w", w.Path, string(out), err)
	}
	return nil
}
