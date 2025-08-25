package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/flaticols/bump/internal/git"
	G "github.com/flaticols/bump/internal/git"
	"github.com/flaticols/bump/internal/tui"
	"github.com/spf13/cobra"
)

// CreateUndoCmd creates a new "undo" command for the CLI, which removes the latest
// semantic version (semver) git tag from both the local and remote repositories.
//
// The command provides the following features:
//   - Prompts the user for confirmation before removing the tag, unless the `--brave`
//     flag is used to bypass confirmation.
//   - Handles errors gracefully, including cases where no tags are found or the latest
//     tag is not a valid semver tag.
//   - Removes the tag locally and optionally from the remote repository, depending on
//     the configuration.
//
// Parameters:
//
//	opts *Options - A pointer to the Options struct containing dependencies and
//	                configuration for the command.
//
// Returns:
//
//	*cobra.Command - The configured "undo" command ready to be added to the CLI.
func CreateUndoCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Remove the latest semver git tag",
		Long:  "Remove the latest semver git tag both locally and from the remote repository",
		Example: "  bump undo           # Removes the latest tag (" +
			"prompts for confirmation)\n  bump undo --brave   # Removes the latest tag without confirmation",
		RunE: func(cmd *cobra.Command, args []string) error {
			ver, err := git.CmdGetTag(opts.Prefix)
			var tagErr G.SemVerTagError
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
			if opts.Prefix != "" {
				tag = opts.Prefix + tag
			}
			confirm := tui.AskConfirmation("Are you sure?", tui.Yes(fmt.Sprintf("Yes remove %s!", tag)), tui.AvoidIf(opts.BraveMode, true))

			if confirm {
				opts.P.Printf("%s removing tag %s\n", opts.P.Symbols.Bullet, opts.P.Info(tag))
				if err := git.CmdRemoveTag(tag); err != nil {
					return err
				}
				opts.P.Printf("%s local tag removed\n", opts.P.Symbols.Ok)
				if !opts.OnlyLocal {
					if err := git.CmdRemoveRemoteTag(tag); err != nil {
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
