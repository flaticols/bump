package cmd

import (
	"errors"
	"fmt"
	"os"
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

type GitStater interface {
	IsDefaultBranch() (string, bool, error)
	CheckLocalChanges() (bool, error)
	CheckRemoteChanges(allowNoRemotes bool) (bool, error)
	HasUnpushedChanges(currentBranch string) (bool, error)
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
}

func CreateRootCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "bump",
		Args:      cobra.OnlyValidArgs,
		ValidArgs: []string{major, minor, patch},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if opts.Verbose {
				fmt.Printf("%s working directory: %s\n", opts.P.Symbols.Bullet, opts.RepoDirectory)
			}

			err := internal.SetBumpWd(opts.RepoDirectory)
			if err != nil {
				fmt.Println(opts.P.Err(err.Error()))
				os.Exit(1)
			}

			if !opts.BraveMode {
				gitStateChecks(opts)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ver, err := opts.GitDetailer.GetCurrentVersion()
			var tagErr internal.SemVerTagError
			var nextVer *semver.Version
			if err != nil {
				if errors.As(err, &tagErr) {
					if !tagErr.NoTags {
						fmt.Println(opts.P.Err("tag '%s' is not a valid semver tag", tagErr.Tag))
						os.Exit(1)
					}

					fmt.Println(opts.P.Warning("no tags found, using default version %s\n", opts.P.Version(internal.DefaultVersion)))
					ver = semver.MustParse("0.0.0")
				} else {
					return err
				}
			}

			nextVer = createNewVersion(getIncPart(args), ver)

			if err != nil && tagErr.NoTags {
				fmt.Printf("set version %s\n", opts.P.Ok("v"+nextVer.String()))
			} else {
				fmt.Printf("%s bump version %s => %s\n", opts.P.Symbols.Bullet, opts.P.Info("v"+ver.String()), opts.P.Ok("v"+nextVer.String()))
			}

			tag := opts.P.Version(nextVer.String())
			err = opts.GitDetailer.SetGitTag(tag)
			if err != nil {
				return err
			}

			if !opts.LocalRepo {
				err = opts.GitDetailer.PushGitTag(tag)
				if err != nil {
					return err
				}
			}

			fmt.Printf("%s %s\n", opts.P.Symbols.Ok, tag)
			return nil
		},
	}

	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.Version = handleVersionCommand()

	return cmd
}

func gitStateChecks(opts *Options) {
	b, yes, err := opts.GitDetailer.IsDefaultBranch()
	if err != nil {
		fmt.Println(opts.P.Err(err.Error()))
		os.Exit(1)
	} else if !yes {
		fmt.Printf("%s not on default branch (%s)\n", opts.P.Symbols.Error, b)
		os.Exit(1)
	} else {
		fmt.Printf("%s on default branch (%s)\n", opts.P.Symbols.Ok, b)
	}

	if yes, err := opts.GitDetailer.CheckLocalChanges(); err != nil {
		fmt.Println(opts.P.Err(err.Error()))
		os.Exit(1)
	} else if yes {
		fmt.Printf("%s uncommitted changes\n", opts.P.Symbols.Error)
		os.Exit(1)
	} else {
		fmt.Printf("%s no uncommitted changes\n", opts.P.Symbols.Ok)
	}

	if yes, err := opts.GitDetailer.CheckRemoteChanges(opts.LocalRepo); err != nil {
		fmt.Println(opts.P.Err(err.Error()))
		os.Exit(1)
	} else if yes {
		fmt.Printf("%s remote changes, pull first\n", opts.P.Symbols.Error)
		os.Exit(1)
	} else {
		fmt.Printf("%s no remote changes\n", opts.P.Symbols.Ok)
	}

	if yes, err := opts.GitDetailer.HasUnpushedChanges(b); err != nil {
		fmt.Println(opts.P.Err(err.Error()))
		os.Exit(1)
	} else if yes {
		fmt.Printf("%s unpushed changes\n", opts.P.Symbols.Error)
		os.Exit(1)
	} else {
		fmt.Printf("%s no unpushed changes\n", opts.P.Symbols.Ok)
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
