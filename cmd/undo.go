package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/flaticols/bump/internal"
	"github.com/flaticols/bump/internal/tui"
	"github.com/spf13/cobra"
)

func CreateUndoCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Remove the latest semver git tag",
		Long:  "Remove the latest semver git tag both locally and from the remote repository",
		Example: "  bump undo           # Removes the latest tag (" +
			"prompts for confirmation)\n  bump undo --brave   # Removes the latest tag without confirmation",
		RunE: func(cmd *cobra.Command, args []string) error {
			ver, err := opts.GitDetailer.GetCurrentVersion()
			var tagErr internal.SemVerTagError
			if err != nil {
				if errors.As(err, &tagErr) {
					if tagErr.NoTags {
						opts.P.Printf("%s no tags found to remove\n", opts.P.Symbols.Error)
						opts.Exit()
					}
					opts.P.Println(opts.P.Err("tag '%s' is not a valid semver tag", tagErr.Tag))
					os.Exit(1)
				}
				return err
			}

			tag := opts.P.Version(ver.String())
			confirm := tui.AskConfirmation("Are you sure?", tui.Yes(fmt.Sprintf("Yes remove %s!", tag)), tui.AvoidIf(opts.BraveMode, true))

			if confirm {
				opts.P.Printf("%s removing tag %s\n", opts.P.Symbols.Bullet, opts.P.Info(tag))
				if err := opts.GitDetailer.RemoveLocalGitTag(tag); err != nil {
					return err
				}
				opts.P.Printf("%s local tag removed\n", opts.P.Symbols.Ok)
				if !opts.LocalRepo {
					if err := opts.GitDetailer.RemoveRemoteGitTag(tag); err != nil {
						opts.P.Printf("%s remote tag not removed\n", opts.P.Symbols.Error)
						opts.P.Printf("%s error: %s\n", opts.P.Symbols.Error, err.Error())
						os.Exit(1)
					}
					opts.P.Printf("%s remote tag removed\n", opts.P.Symbols.Ok)
				}
			}

			return nil
		},
	}

	return cmd
}
