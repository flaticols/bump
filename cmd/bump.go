package cmd

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/flaticols/bump/internal"
	"github.com/spf13/cobra"
	"os"
	"path"
	"runtime/debug"
)

type semVerPart = string

const (
	major semVerPart = "major"
	minor semVerPart = "minor"
	patch semVerPart = "patch"
)

type ColorTextPrinter func(format string, a ...interface{}) string

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
	cmd := &cobra.Command{
		Use:       "bump",
		Args:      cobra.OnlyValidArgs,
		ValidArgs: []string{major, minor, patch},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			verbose, _ := cmd.Flags().GetBool("verbose")
			repoPath, _ := cmd.PersistentFlags().GetString("repo")
			allowNoRemotes, _ := cmd.PersistentFlags().GetBool("no-remotes")
			if repoPath == "" {
				wd, err := os.Getwd()
				if err != nil {
					fmt.Println(opts.ErrPrinter(err.Error()))
					os.Exit(1)
				}
				repoPath = wd
			}

			if verbose {
				fmt.Printf("Working directory: %s\n", repoPath)
			}

			err := os.Chdir(path.Clean(repoPath))
			if err != nil {
				fmt.Println(opts.ErrPrinter(err.Error()))
				os.Exit(1)
			}

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

					fmt.Println(opts.WarningPrinter("no tags found, using default version %s\n", internal.DefaultVersion))
					ver = semver.MustParse(internal.DefaultVersion)
					nextVer = semver.MustParse(internal.DefaultVersion)
				} else {
					return err
				}
			}

			incPart := patch

			if len(args) > 0 {
				incPart = args[0]
			}

			switch incPart {
			case major:
				v := ver.IncMajor()
				nextVer = &v
			case minor:
				v := ver.IncMinor()
				nextVer = &v
			case patch:
				fallthrough
			default:
				v := ver.IncPatch()
				nextVer = &v
			}

			if nextVer.GreaterThan(ver) {
				fmt.Printf("bump version %s => %s\n", opts.InfoPrinter(ver.String()), opts.OkPrinter(nextVer.String()))
			} else {
				fmt.Printf("set version %s\n", opts.OkPrinter(nextVer.String()))
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

	cmd.PersistentFlags().String("repo", "", "path to the repository")
	cmd.PersistentFlags().Bool("verbose", false, "enable verbose output")
	cmd.PersistentFlags().Bool("no-remotes", false, "if no-remotes is set, bump will not error if no remotes are found")

	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.Version = handleVersionCommand()

	return cmd
}

// handleVersionCommand handles the version command and exits.
func handleVersionCommand() string {
	info, _ := debug.ReadBuildInfo()
	return info.Main.Version
}
