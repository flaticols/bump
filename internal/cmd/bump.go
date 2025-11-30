package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/flaticols/bump/internal/git"
	G "github.com/flaticols/bump/internal/git"
	"github.com/flaticols/bump/semver"
	"github.com/spf13/cobra"
)

type semVerPart = string

const (
	major semVerPart = "major"
	minor semVerPart = "minor"
	patch semVerPart = "patch"
)

type (
	ColorTextPrinter func(format string, a ...any) string
	VersionPrinter   func(string) string
	Printf           func(format string, a ...any)
	Println          func(format string, a ...any)
)

type Symbols struct {
	Ok      string
	Warning string
	Error   string
	Bullet  string
}

type TextPrinters struct {
	Err     ColorTextPrinter
	Info    ColorTextPrinter
	Warning ColorTextPrinter
	Ok      ColorTextPrinter
	Printf  Printf
	Println Println
	Version VersionPrinter
	Symbols Symbols
}

type JSONTag struct {
	Current string `json:"current"`
	New     string `json:"new"`
}

type JSONResult struct {
	Successful bool     `json:"successful"`
	Checks     []string `json:"checks"`
	Tag        JSONTag  `json:"tag"`
}

type Options struct {
	Exit           func()
	P              TextPrinters
	RepoDirectory  string
	DefaultBranchs []string
	Verbose        bool
	OnlyLocal      bool
	BraveMode      bool
	NoColor        bool
	JSON           bool
	Prefix         string
	Result         JSONResult
}

// CreateRootCmd initializes and returns the root command for the "bump" CLI tool.
// This command provides functionality to increment semantic versioning tags in Git repositories.
//
// Parameters:
//   - opts: A pointer to an Options struct that contains configuration and dependencies for the command.
//
// Returns:
//   - *cobra.Command: The root command for the "bump" CLI tool.
//
// The command supports the following subcommands:
//   - major: Increments the major version (e.g., v1.2.3 -> v2.0.0).
//   - minor: Increments the minor version (e.g., v1.2.3 -> v1.3.0).
//   - patch: Increments the patch version (e.g., v1.2.3 -> v1.2.4).
//
// Features:
//   - Automatically detects the current semantic version tag in the Git repository.
//   - Handles cases where no tags are present by using a default version (0.0.0).
//   - Validates semantic versioning tags and provides error messages for invalid tags.
//   - Creates and optionally pushes the new version tag to the remote repository.
//
// Example Usage:
//   - bump         # Bumps the patch version (e.g., v1.2.3 -> v1.2.4).
//   - bump major   # Bumps the major version (e.g., v1.2.3 -> v2.0.0).
//   - bump minor   # Bumps the minor version (e.g., v1.2.3 -> v1.3.0).
//   - bump patch   # Bumps the patch version (e.g., v1.2.3 -> v1.2.4).
func CreateRootCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bump [major|minor|patch] [package]",
		Short:   "A command-line tool to easily bump the git tag version of your project using semantic versioning",
		Long:    `Bump is a lightweight command-line tool that helps you manage semantic versioning tags in Git repositories. It automates version increments following SemVer standards, making it easy to maintain proper versioning in your projects.`,
		Example: "  bump              # Bumps patch version (e.g., v1.2.3 -> v1.2.4)\n  bump major        # Bumps major version (e.g., v1.2.3 -> v2.0.0)\n  bump minor        # Bumps minor version (e.g., v1.2.3 -> v1.3.0)\n  bump patch        # Bumps patch version (e.g., v1.2.3 -> v1.2.4)\n  bump pkg/x        # Bumps patch for pkg/x (e.g., pkg/x/v1.2.3 -> pkg/x/v1.2.4)\n  bump major pkg/x  # Bumps major for pkg/x (e.g., pkg/x/v1.2.3 -> pkg/x/v2.0.0)",
		Args:    cobra.MaximumNArgs(2),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return gitStateChecks(opts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse arguments and normalize prefix
			incPart := getIncPart(args)
			opts.Prefix = normalizePrefix(getPackageName(args), opts.Prefix)

			// Get current version or start from 0.0.0
			ver, noTags, err := getCurrentVersion(opts)
			if err != nil {
				return err
			}

			// Store current tag in result
			if !noTags {
				opts.Result.Tag.Current = formatTag(ver.String(), opts.Prefix, opts.P.Version)
			}

			// Calculate and create new version
			nextVer := createNewVersion(incPart, ver)
			newTag := formatTag(nextVer.String(), opts.Prefix, opts.P.Version)
			opts.Result.Tag.New = newTag

			// Print version change
			printVersionChange(opts, noTags, ver.String(), newTag)

			// Create and push tag
			if err := git.CmdCreateTag(newTag); err != nil {
				return err
			}
			opts.P.Printf("%s tag %s created\n", opts.P.Symbols.Ok, newTag)

			if !opts.OnlyLocal {
				if err := git.CmdPushTag(newTag); err != nil {
					return err
				}
				opts.P.Printf("%s tag %s pushed\n", opts.P.Symbols.Ok, newTag)
			}

			opts.Result.Successful = true
			return nil
		},
	}

	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.Version = handleVersionCommand()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	return cmd
}

// getCurrentVersion retrieves the current version from git tags or returns 0.0.0 if no tags exist
func getCurrentVersion(opts *Options) (semver.Version, bool, error) {
	ver, err := git.CmdGetTag(opts.Prefix)
	if err != nil {
		var tagErr G.SemVerTagError
		if errors.As(err, &tagErr) {
			if !tagErr.NoTags {
				return semver.Version{}, false, fmt.Errorf("invalid semver tag: %s", tagErr.Tag)
			}
			// No tags: start from 0.0.0
			opts.P.Printf("%s no tags found, using default version %s\n",
				opts.P.Symbols.Bullet, opts.P.Version(G.DefaultVersion))
			ver, _ = semver.Parse("0.0.0")
			return ver, true, nil
		}
		return semver.Version{}, false, err
	}
	return ver, false, nil
}

// normalizePrefix ensures the prefix has a trailing slash if non-empty
// Positional package argument takes priority over --prefix flag
func normalizePrefix(pkgName, prefixFlag string) string {
	prefix := prefixFlag
	if pkgName != "" {
		prefix = pkgName
	}

	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		return prefix + "/"
	}
	return prefix
}

// formatTag formats a version string with the prefix and version formatter
func formatTag(version, prefix string, formatter VersionPrinter) string {
	tag := formatter(version)
	if prefix != "" {
		return prefix + tag
	}
	return tag
}

// printVersionChange prints the appropriate version change message
func printVersionChange(opts *Options, noTags bool, oldVersion, newTag string) {
	if noTags {
		opts.P.Printf("%s set tag %s\n", opts.P.Symbols.Ok, newTag)
	} else {
		prevTag := formatTag(oldVersion, opts.Prefix, opts.P.Version)
		opts.P.Printf("%s bump tag %s => %s\n", opts.P.Symbols.Bullet, prevTag, newTag)
	}
}

// gitStateChecks performs a series of checks on the Git repository state to ensure
// it is in a valid state for further operations. The checks include:
//
// 1. Verifying if the current branch is the default branch.
// 2. Checking for uncommitted changes in the working directory.
// 3. Checking for remote changes that need to be pulled.
// 4. Checking for unpushed changes in the local repository.
// 5. Checking for unfetched remote tags (if the repository is not local).
//
// If any of these checks fail, the function will print an appropriate error or
// warning message and exit the program unless BraveMode is enabled in the options.
//
// Parameters:
//   - opts (*Options): A pointer to an Options struct containing configuration
//     and utility methods for performing Git operations and printing messages.
func gitStateChecks(opts *Options) error {
	// Check current branch
	branch, err := git.CmdCurrentBranch()
	if err := handleGitError(opts, err); err != nil {
		return err
	}

	if !slices.Contains(opts.DefaultBranchs, branch) {
		if err := handleCheckFailure(opts, fmt.Sprintf("not on default branch (%s)", branch), "not on default branch"); err != nil {
			return err
		}
	} else {
		opts.P.Printf("%s on default branch (%s)\n", opts.P.Symbols.Ok, branch)
	}

	// Check local changes
	if err := checkCondition(opts, git.CmdHasLocalChanges,
		"uncommitted changes", "uncommitted changes",
		"no uncommitted changes"); err != nil {
		return err
	}

	// Remote checks
	if !opts.OnlyLocal {
		if err := checkCondition(opts, git.CmdHasRemoteChanges,
			"remote changes, pull first", "remote changes present",
			"no remote changes"); err != nil {
			return err
		}

		if err := checkCondition(opts, func() (bool, error) { return git.CmdHasUnpushedChanges(branch) },
			"unpushed changes", "unpushed changes present",
			"no unpushed changes"); err != nil {
			return err
		}

		if err := handleRemoteTags(opts); err != nil {
			return err
		}
	}

	return nil
}

// handleGitError handles git command errors with brave mode support
func handleGitError(opts *Options, err error) error {
	if err == nil {
		return nil
	}
	opts.P.Printf("%s %s\n", opts.P.Symbols.Error, err.Error())
	if opts.BraveMode {
		return nil
	}
	return err
}

// handleCheckFailure handles check failures with brave mode support
func handleCheckFailure(opts *Options, message, errMsg string) error {
	opts.P.Printf("%s %s\n", opts.P.Symbols.Error, message)
	if opts.BraveMode {
		return nil
	}
	return errors.New(errMsg)
}

// checkCondition runs a boolean check and handles the result
func checkCondition(opts *Options, check func() (bool, error), failMsg, errMsg, okMsg string) error {
	result, err := check()
	if err != nil {
		return handleGitError(opts, err)
	}
	if result {
		return handleCheckFailure(opts, failMsg, errMsg)
	}
	opts.P.Printf("%s %s\n", opts.P.Symbols.Ok, okMsg)
	return nil
}

// handleRemoteTags checks and fetches remote tags if needed
func handleRemoteTags(opts *Options) error {
	yes, err := git.CmdHasRemoteUnfetchedTags()
	if err != nil {
		opts.P.Printf("%s %s\n", opts.P.Symbols.Warning, err.Error())
		return nil // Non-fatal
	}

	if !yes {
		opts.P.Printf("%s no new remote tags\n", opts.P.Symbols.Ok)
		return nil
	}

	opts.P.Printf("%s remote has new tags, fetching tags first\n", opts.P.Symbols.Warning)
	fetchCmd := exec.Command("git", "fetch", "--tags")
	if err := fetchCmd.Run(); err != nil {
		opts.P.Printf("%s failed to fetch tags: %s\n", opts.P.Symbols.Error, err.Error())
		if !opts.BraveMode {
			return err
		}
	}
	opts.P.Printf("%s tags fetched successfully\n", opts.P.Symbols.Ok)
	return nil
}

// handleVersionCommand handles the version command and exits.
func handleVersionCommand() string {
	info, _ := debug.ReadBuildInfo()
	return info.Main.Version
}

// getIncPart returns the semantic version part to increment based on the provided arguments.
// If the first argument is a valid version part (major/minor/patch), it is returned.
// Otherwise, it defaults to returning patch.
func getIncPart(args []string) semVerPart {
	if len(args) > 0 {
		firstArg := args[0]
		// Check if first arg is a valid version part
		if firstArg == major || firstArg == minor || firstArg == patch {
			return firstArg
		}
		// First arg is not a version part, so it must be a package name
		// Default to patch
		return patch
	}
	return patch
}

// getPackageName extracts the package name from command arguments.
// Returns empty string if no package name is provided.
// Supports two argument patterns:
//   - bump <package>           # package is first arg
//   - bump <version> <package> # package is second arg
func getPackageName(args []string) string {
	if len(args) == 0 {
		return ""
	}

	firstArg := args[0]

	// If we have 2 args, second is always the package name
	if len(args) == 2 {
		return args[1]
	}

	// If we have 1 arg and it's a version part, no package name
	if firstArg == major || firstArg == minor || firstArg == patch {
		return ""
	}

	// Single arg that's not a version part must be the package name
	return firstArg
}

// createNewVersion returns a new semantic version by incrementing the specified part of the provided version.
// The incPart parameter determines which part of the version to increment: major for a major update,
// minor for a minor update, and patch (or any unrecognized value, due to fallthrough) for a patch update.
// It returns the updated semver.Version after performing the corresponding increment operation.
func createNewVersion(incPart semVerPart, ver semver.Version) semver.Version {
	switch incPart {
	case major:
		return ver.IncrementMajor()
	case minor:
		return ver.IncrementMinor()
	case patch:
		fallthrough
	default:
		return ver.IncrementPatch()
	}
}
