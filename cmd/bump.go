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

type GitStater interface {
	IsDefaultBranch() (string, bool, error)
	CheckLocalChanges() (bool, error)
	CheckRemoteChanges(allowNoRemotes bool) (bool, error)
	HasUnpushedChanges(currentBranch string) (bool, error)
	GetCurrentVersion() (*semver.Version, error)
	SetGitTag(tag string) error
	PushGitTag(tag string) error
}

type Options struct {
	ErrPrinter     ColorTextPrinter
	InfoPrinter    ColorTextPrinter
	WarningPrinter ColorTextPrinter
	OkPrinter      ColorTextPrinter
	GitDetailer    GitStater
}

func CreateRootCmd(opts *Options) *cobra.Command {
	var repoDirectory string
	var verbose bool
	var allowNoRemotes bool
	var strictMode bool
	cmd := &cobra.Command{
		Use:       "bump",
		Args:      cobra.OnlyValidArgs,
		ValidArgs: []string{major, minor, patch},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				fmt.Printf("working directory: %s\n", repoDirectory)
			}

			err := internal.SetBumpWd(repoDirectory)
			if err != nil {
				fmt.Println(opts.ErrPrinter(err.Error()))
				os.Exit(1)
			}

			if strictMode {
				gitStateChecks(opts, allowNoRemotes)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			allowNoRemotes, _ := cmd.PersistentFlags().GetBool("no-remotes")

			ver, err := opts.GitDetailer.GetCurrentVersion()
			var tagErr internal.SemVerTagError
			var nextVer *semver.Version
			if err != nil {
				if errors.As(err, &tagErr) {
					if !tagErr.NoTags {
						fmt.Println(opts.ErrPrinter("tag '%s' is not a valid semver tag", tagErr.Tag))
						os.Exit(1)
					}

					fmt.Println(opts.WarningPrinter("no tags found, using default version v%s\n", internal.DefaultVersion))
					ver = semver.MustParse("0.0.0")
				} else {
					return err
				}
			}

			nextVer = createNewVersion(getIncPart(args), ver)

			if err != nil && tagErr.NoTags {
				fmt.Printf("set version %s\n", opts.OkPrinter("v"+nextVer.String()))
			} else {
				fmt.Printf("bump version %s => %s\n", opts.InfoPrinter("v"+ver.String()), opts.OkPrinter("v"+nextVer.String()))
			}

			tag := fmt.Sprintf("v%s", nextVer.String())
			err = opts.GitDetailer.SetGitTag(tag)
			if err != nil {
				return err
			}

			if !allowNoRemotes {
				err = opts.GitDetailer.PushGitTag(tag)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&repoDirectory, "repo", "r", "", "path to the repository")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	cmd.PersistentFlags().BoolVarP(&allowNoRemotes, "no-remotes", "l", false, "if no-remotes is set, bump will not error if no remotes are found")
	cmd.PersistentFlags().BoolVarP(&strictMode, "strict", "s", true, "if strict is set, bump will error if git state checks are not met (default: true)")

	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.Version = handleVersionCommand()

	undoCmd := CreateUndoCmd(opts)
	cmd.AddCommand(undoCmd)

	return cmd
}

func gitStateChecks(opts *Options, allowNoRemotes bool) {
	b, yes, err := opts.GitDetailer.IsDefaultBranch()
	if err != nil {
		fmt.Println(opts.ErrPrinter(err.Error()))
		os.Exit(1)
	} else if !yes {
		fmt.Println(opts.ErrPrinter("not on default branch"))
		os.Exit(1)
	} else {
		fmt.Println(opts.OkPrinter("on default branch: %s", b))
	}

	if yes, err := opts.GitDetailer.CheckLocalChanges(); err != nil {
		fmt.Println(opts.ErrPrinter(err.Error()))
		os.Exit(1)
	} else if yes {
		fmt.Println(opts.ErrPrinter("uncommitted changes"))
		os.Exit(1)
	} else {
		fmt.Println(opts.OkPrinter("no uncommitted changes"))
	}

	if yes, err := opts.GitDetailer.CheckRemoteChanges(allowNoRemotes); err != nil {
		fmt.Println(opts.ErrPrinter(err.Error()))
		os.Exit(1)
	} else if yes {
		fmt.Println(opts.ErrPrinter("remote changes, pull first"))
		os.Exit(1)
	} else {
		fmt.Println(opts.OkPrinter("no remote changes"))
	}

	if yes, err := opts.GitDetailer.HasUnpushedChanges(b); err != nil {
		fmt.Println(opts.ErrPrinter(err.Error()))
		os.Exit(1)
	} else if yes {
		fmt.Println(opts.ErrPrinter("unpushed changes"))
		os.Exit(1)
	} else {
		fmt.Println(opts.OkPrinter("no unpushed changes"))
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
