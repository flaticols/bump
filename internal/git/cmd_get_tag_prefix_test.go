package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// helper to run a git command in a specific directory with optional env vars
func runGit(t *testing.T, dir string, env []string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\noutput: %s", args, err, string(out))
	}
	return string(out)
}

// create a temp git repo with an initial commit
func newTempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, nil, "init")
	runGit(t, dir, nil, "config", "user.email", "test@example.com")
	runGit(t, dir, nil, "config", "user.name", "Test User")
	// create a file and commit
	content := []byte("hello")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), content, 0o644))
	runGit(t, dir, nil, "add", ".")
	runGit(t, dir, nil, "commit", "-m", "init")
	return dir
}

// create annotated tag with the given tagger date (ISO 8601 acceptable by git), ensuring same creatordate grouping
func createAnnotatedTagWithDate(t *testing.T, dir, tag, isoDate string) {
	t.Helper()
	env := []string{
		"GIT_COMMITTER_DATE=" + isoDate,
		"GIT_AUTHOR_DATE=" + isoDate,
	}
	runGit(t, dir, env, "tag", "-a", tag, "-m", "test")
}

func TestCmdGetTag_NoPrefix_UsesLatestGroupHighestSemver(t *testing.T) {
	repo := newTempRepo(t)

	// Older group (2000) with a higher version number to ensure grouping logic is respected
	createAnnotatedTagWithDate(t, repo, "v2.0.0", "2000-01-01T00:00:00Z")

	// Latest group (2001) with two tags
	createAnnotatedTagWithDate(t, repo, "v1.0.0", "2001-01-01T00:00:00Z")
	createAnnotatedTagWithDate(t, repo, "v1.1.0", "2001-01-01T00:00:00Z")

	// Ensure we run CmdGetTag in the repo directory
	cwd, _ := os.Getwd()
	require.NoError(t, os.Chdir(repo))
	defer func() { _ = os.Chdir(cwd) }()

	ver, err := CmdGetTag("")
	require.NoError(t, err)
	require.Equal(t, "1.1.0", ver.String())
}

func TestCmdGetTag_WithPrefix_FiltersAndStripsPrefix(t *testing.T) {
	repo := newTempRepo(t)

	// Newest group (2002) contains non-prefixed tag only
	createAnnotatedTagWithDate(t, repo, "v9.9.9", "2002-01-01T00:00:00Z")

	// Next group (2001) contains prefixed tags
	createAnnotatedTagWithDate(t, repo, "pkg/x/v0.2.0", "2001-01-01T00:00:00Z")
	createAnnotatedTagWithDate(t, repo, "pkg/x/v0.3.0", "2001-01-01T00:00:00Z")

	cwd, _ := os.Getwd()
	require.NoError(t, os.Chdir(repo))
	defer func() { _ = os.Chdir(cwd) }()

	ver, err := CmdGetTag("pkg/x/")
	require.NoError(t, err)
	require.Equal(t, "0.3.0", ver.String())
}

func TestCmdGetTag_WithPrefix_NoMatchingTagsReturnsNoTags(t *testing.T) {
	repo := newTempRepo(t)
	createAnnotatedTagWithDate(t, repo, "v1.2.3", "2001-01-01T00:00:00Z")

	cwd, _ := os.Getwd()
	require.NoError(t, os.Chdir(repo))
	defer func() { _ = os.Chdir(cwd) }()

	_, err := CmdGetTag("pkg/x/")
	var tagErr SemVerTagError
	require.Error(t, err)
	require.True(t, errors.As(err, &tagErr))
	require.True(t, tagErr.NoTags)
}
