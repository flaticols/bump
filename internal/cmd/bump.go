package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"slices"

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
		Use:       "bump [major|minor|patch]",
		Short:     "A command-line tool to easily bump the git tag version of your project using semantic versioning",
		Long:      `Bump is a lightweight command-line tool that helps you manage semantic versioning tags in Git repositories. It automates version increments following SemVer standards, making it easy to maintain proper versioning in your projects.`,
		Example:   "  bump         # Bumps patch version (e.g., v1.2.3 -> v1.2.4)\n  bump major   # Bumps major version (e.g., v1.2.3 -> v2.0.0)\n  bump minor   # Bumps minor version (e.g., v1.2.3 -> v1.3.0)\n  bump patch   # Bumps patch version (e.g., v1.2.3 -> v1.2.4)",
		Args:      cobra.OnlyValidArgs,
		ValidArgs: []string{major, minor, patch},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return gitStateChecks(opts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ver, err := git.CmdGetTag()
			var tagErr G.SemVerTagError
			var nextVer semver.Version
			if err != nil {
				if errors.As(err, &tagErr) {
					if !tagErr.NoTags {
						return fmt.Errorf("invalid semver tag: %s", tagErr.Tag)
					}
					// No tags: start from 0.0.0
					opts.P.Printf("%s no tags found, using default version %s\n", opts.P.Symbols.Bullet,
						opts.P.Version(G.DefaultVersion))
					ver, _ = semver.Parse("0.0.0")
				} else {
					return err
				}
			}
			
			// Persist current tag (empty if NoTags)
			if err == nil {
				opts.Result.Tag.Current = opts.P.Version(ver.String())
			} else if tagErr.NoTags {
				opts.Result.Tag.Current = ""
			}
			
			nextVer = createNewVersion(getIncPart(args), ver)
			tag := opts.P.Version(nextVer.String())
			opts.Result.Tag.New = tag
			
			if err != nil && tagErr.NoTags {
				opts.P.Printf("%s set tag %s\n", opts.P.Symbols.Ok, tag)
			} else {
				opts.P.Printf("%s bump tag %s => %s\n", opts.P.Symbols.Bullet, opts.P.Version(ver.String()), tag)
			}
			
			err = git.CmdCreateTag(tag)
			if err != nil {
				return err
			}
			opts.P.Printf("%s tag %s created\n", opts.P.Symbols.Ok, tag)
			
			if !opts.OnlyLocal {
				err = git.CmdPushTag(tag)
				if err != nil {
					return err
				}
				opts.P.Printf("%s tag %s pushed\n", opts.P.Symbols.Ok, tag)
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
	branch, err := git.CmdCurrentBranch()
	if err != nil {
		opts.P.Printf("%s %s\n", opts.P.Symbols.Error, err.Error())
		if !opts.BraveMode {
			return err
		}
	}
	ok := slices.Contains(opts.DefaultBranchs, branch)
	if !ok {
		opts.P.Printf("%s not on default branch (%s)\n", opts.P.Symbols.Error, branch)
		if !opts.BraveMode {
			return errors.New("not on default branch")
		}
	} else {
		opts.P.Printf("%s on default branch (%s)\n", opts.P.Symbols.Ok, branch)
	}

	if yes, err := git.CmdHasLocalChanges(); err != nil {
		opts.P.Printf("%s %s\n", opts.P.Symbols.Error, err.Error())
		if !opts.BraveMode {
			return err
		}
	} else if yes {
		opts.P.Printf("%s uncommitted changes\n", opts.P.Symbols.Error)
		if !opts.BraveMode {
			return errors.New("uncommitted changes")
		}
	} else {
		opts.P.Printf("%s no uncommitted changes\n", opts.P.Symbols.Ok)
	}

	if !opts.OnlyLocal {
		if yes, err := git.CmdHasRemoteChanges(); err != nil {
			opts.P.Printf("%s %s\n", opts.P.Symbols.Error, err.Error())
			if !opts.BraveMode {
				return err
			}
		} else if yes {
			opts.P.Printf("%s remote changes, pull first\n", opts.P.Symbols.Error)
			if !opts.BraveMode {
				return errors.New("remote changes present")
			}
		} else {
			opts.P.Printf("%s no remote changes\n", opts.P.Symbols.Ok)
		}

		if yes, err := git.CmdHasUnpushedChanges(branch); err != nil {
			opts.P.Printf("%s %s\n", opts.P.Symbols.Error, err.Error())
			if !opts.BraveMode {
				return err
			}
		} else if yes {
			opts.P.Printf("%s unpushed changes\n", opts.P.Symbols.Error)
			if !opts.BraveMode {
				return errors.New("unpushed changes present")
			}
		} else {
			opts.P.Printf("%s no unpushed changes\n", opts.P.Symbols.Ok)
		}

		if yes, err := git.CmdHasRemoteUnfetchedTags(); err != nil {
			opts.P.Printf("%s %s\n", opts.P.Symbols.Warning, err.Error())
		} else if yes {
			opts.P.Printf("%s remote has new tags, fetching tags first\n", opts.P.Symbols.Warning)
			fetchCmd := exec.Command("git", "fetch", "--tags")
			if err := fetchCmd.Run(); err != nil {
				opts.P.Printf("%s failed to fetch tags: %s\n", opts.P.Symbols.Error, err.Error())
				if !opts.BraveMode {
					return err
				}
			}
			opts.P.Printf("%s tags fetched successfully\n", opts.P.Symbols.Ok)
		} else {
			opts.P.Printf("%s no new remote tags\n", opts.P.Symbols.Ok)
		}
	}
	return nil
}

// handleVersionCommand handles the version command and exits.
func handleVersionCommand() string {
	info, _ := debug.ReadBuildInfo()
	return info.Main.Version
}

// getIncPart returns the semantic version part to increment based on the provided arguments.
// If the input slice contains at least one element, the first element is returned as the part to increment.
// Otherwise, it defaults to returning the patch part.
func getIncPart(args []string) semVerPart {
	if len(args) > 0 {
		return args[0]
	}
	return patch
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
