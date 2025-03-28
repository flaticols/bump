package cmd

import (
	"errors"
	"os"
	"os/exec"
	"runtime/debug"

	"github.com/Masterminds/semver/v3"
	"github.com/flaticols/bump/internal"
	"github.com/spf13/cobra"
)

type semVerPart = string

const (
	major semVerPart = "major"
	minor semVerPart = "minor"
	patch semVerPart = "patch"
)

type ColorTextPrinter func(format string, a ...any) string
type VersionPrinter func(string) string
type Printf func(format string, a ...any)
type Println func(format string, a ...any)

type GitStater interface {
	IsDefaultBranch() (string, bool, error)
	CheckLocalChanges() (bool, error)
	CheckRemoteChanges(allowNoRemotes bool) (bool, error)
	HasUnpushedChanges(currentBranch string) (bool, error)
	HasRemoteUnfetchedTags() (bool, error)
	GetCurrentVersion() (*semver.Version, error)
	SetGitTag(string) error
	PushGitTag(string) error
	RemoveLocalGitTag(string) error
	RemoveRemoteGitTag(string) error
}

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

type Options struct {
	P                  TextPrinters
	GitDetailer        GitStater
	RepoDirectory      string
	Verbose, LocalRepo bool
	BraveMode          bool //ignore any warning just try to do all the things
	NoColor            bool
	Exit               func()
}

func CreateRootCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "bump [major|minor|patch]",
		Short:     "A command-line tool to easily bump the git tag version of your project using semantic versioning",
		Long:      `Bump is a lightweight command-line tool that helps you manage semantic versioning tags in Git repositories. It automates version increments following SemVer standards, making it easy to maintain proper versioning in your projects.`,
		Example:   "  bump         # Bumps patch version (e.g., v1.2.3 -> v1.2.4)\n  bump major   # Bumps major version (e.g., v1.2.3 -> v2.0.0)\n  bump minor   # Bumps minor version (e.g., v1.2.3 -> v1.3.0)\n  bump patch   # Bumps patch version (e.g., v1.2.3 -> v1.2.4)",
		Args:      cobra.OnlyValidArgs,
		ValidArgs: []string{major, minor, patch},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
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

			gitStateChecks(opts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ver, err := opts.GitDetailer.GetCurrentVersion()
			var tagErr internal.SemVerTagError
			var nextVer *semver.Version
			if err != nil {
				if errors.As(err, &tagErr) {
					if !tagErr.NoTags {
						opts.P.Println(opts.P.Err("tag '%s' is not a valid semver tag", tagErr.Tag))
						os.Exit(1)
					}

					opts.P.Printf("%s no tags found, using default version %s\n", opts.P.Symbols.Bullet, opts.P.Version(internal.DefaultVersion))
					ver = semver.MustParse("0.0.0")
				} else {
					return err
				}
			}

			nextVer = createNewVersion(getIncPart(args), ver)
			tag := opts.P.Version(nextVer.String())

			if err != nil && tagErr.NoTags {
				opts.P.Printf("%s set tag %s\n", opts.P.Symbols.Ok, tag)
			} else {
				opts.P.Printf("%s bump tag %s => %s\n", opts.P.Symbols.Bullet, opts.P.Version(ver.String()), tag)
			}

			err = opts.GitDetailer.SetGitTag(tag)
			if err != nil {
				return err
			}
			opts.P.Printf("%s tag %s created\n", opts.P.Symbols.Ok, tag)

			if !opts.LocalRepo {
				err = opts.GitDetailer.PushGitTag(tag)
				if err != nil {
					return err
				}
				opts.P.Printf("%s tag %s pushed\n", opts.P.Symbols.Ok, tag)
			}

			return nil
		},
	}

	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.Version = handleVersionCommand()

	return cmd
}

func gitStateChecks(opts *Options) {
	exitIfNotBrave := func() {
		if !opts.BraveMode {
			os.Exit(1)
		}
	}

	b, yes, err := opts.GitDetailer.IsDefaultBranch()
	if err != nil {
		opts.P.Printf("%s %s\n", opts.P.Symbols.Error, err.Error())
		exitIfNotBrave()
	} else if !yes {
		opts.P.Printf("%s not on default branch (%s)\n", opts.P.Symbols.Error, b)
		exitIfNotBrave()
	} else {
		opts.P.Printf("%s on default branch (%s)\n", opts.P.Symbols.Ok, b)
	}

	if yes, err := opts.GitDetailer.CheckLocalChanges(); err != nil {
		opts.P.Printf("%s %s\n", opts.P.Symbols.Error, err.Error())
		exitIfNotBrave()
	} else if yes {
		opts.P.Printf("%s uncommitted changes\n", opts.P.Symbols.Error)
		exitIfNotBrave()
	} else {
		opts.P.Printf("%s no uncommitted changes\n", opts.P.Symbols.Ok)
	}

	if yes, err := opts.GitDetailer.CheckRemoteChanges(opts.LocalRepo); err != nil {
		opts.P.Printf("%s %s\n", opts.P.Symbols.Error, err.Error())
		exitIfNotBrave()
	} else if yes {
		opts.P.Printf("%s remote changes, pull first\n", opts.P.Symbols.Error)
		exitIfNotBrave()
	} else {
		opts.P.Printf("%s no remote changes\n", opts.P.Symbols.Ok)
	}

	if yes, err := opts.GitDetailer.HasUnpushedChanges(b); err != nil {
		opts.P.Printf("%s %s\n", opts.P.Symbols.Error, err.Error())
		exitIfNotBrave()
	} else if yes {
		opts.P.Printf("%s unpushed changes\n", opts.P.Symbols.Error)
		exitIfNotBrave()
	} else {
		opts.P.Printf("%s no unpushed changes\n", opts.P.Symbols.Ok)
	}

	// Check for unfetched remote tags
	if !opts.LocalRepo {
		if yes, err := opts.GitDetailer.HasRemoteUnfetchedTags(); err != nil {
			opts.P.Printf("%s %s\n", opts.P.Symbols.Warning, err.Error())
		} else if yes {
			opts.P.Printf("%s remote has new tags, fetching tags first\n", opts.P.Symbols.Warning)
			fetchCmd := exec.Command("git", "fetch", "--tags")
			if err := fetchCmd.Run(); err != nil {
				opts.P.Printf("%s failed to fetch tags: %s\n", opts.P.Symbols.Error, err.Error())
				exitIfNotBrave()
			}
			opts.P.Printf("%s tags fetched successfully\n", opts.P.Symbols.Ok)
		} else {
			opts.P.Printf("%s no new remote tags\n", opts.P.Symbols.Ok)
		}
	}
}

// handleVersionCommand handles the version command and exits.
func handleVersionCommand() string {
	info, _ := debug.ReadBuildInfo()
	return info.Main.Version
}

func getIncPart(args []string) semVerPart {
	if len(args) > 0 {
		return args[0]
	}
	return patch
}

func createNewVersion(incPart semVerPart, ver *semver.Version) *semver.Version {
	switch incPart {
	case major:
		v := ver.IncMajor()
		return &v
	case minor:
		v := ver.IncMinor()
		return &v
	case patch:
		fallthrough
	default:
		v := ver.IncPatch()
		return &v
	}
}
