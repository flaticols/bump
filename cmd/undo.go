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
		RunE: func(cmd *cobra.Command, args []string) error {
			ver, err := opts.GitDetailer.GetCurrentVersion()
			var tagErr internal.SemVerTagError
			if err != nil {
				if errors.As(err, &tagErr) {
					if tagErr.NoTags {
						fmt.Println(opts.P.Err("no tags found to remove"))
						os.Exit(1)
					}
					fmt.Println(opts.P.Err("tag '%s' is not a valid semver tag", tagErr.Tag))
					os.Exit(1)
				}
				return err
			}

			tag := opts.P.Version(ver.String())
			confirm := tui.AskConfirmation("Are you sure?", tui.Yes(fmt.Sprintf("Yes remove %s!", tag)), tui.AvoidIf(opts.BraveMode, true))

			if confirm {
				fmt.Printf("%s removing tag %s\n", opts.P.Symbols.Bullet, opts.P.Info(tag))
				if err := opts.GitDetailer.RemoveLocalGitTag(tag); err != nil {
					return err
				}
				fmt.Printf("%s local tag removed\n", opts.P.Symbols.Ok)
				if !opts.LocalRepo {
					if err := opts.GitDetailer.RemoveRemoteGitTag(tag); err != nil {
						fmt.Println(opts.P.Warning("failed to remove remote tag: %v", err))
						fmt.Println(opts.P.Warning("local tag was removed, but remote tag may still exist"))
						return err
					}
					fmt.Println(opts.P.Ok("remote tag removed"))
				}
			}

			return nil
		},
	}

	return cmd
}
