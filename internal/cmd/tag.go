package cmd

import (
	"fmt"

	"github.com/flaticols/bump/internal/git"
	"github.com/flaticols/bump/semver"
	"github.com/spf13/cobra"
)

// CreateTagCmd creates the `tag` subcommand which creates a tag from an explicit version.
func CreateTagCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag <version>",
		Short: "Create a tag from an explicit semantic version",
		Long:  "Create a new git tag using the provided semantic version without incrementing the previous one.",
		Args:  cobra.ExactArgs(1),
		Example: "  bump tag 2.3.4         # Creates tag v2.3.4 (or 2.3.4 with --no-v-prefix)\n  bump tag v1.0.0       # 'v' prefix in input is accepted",
		RunE: func(cmd *cobra.Command, args []string) error {
			ver, err := semver.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid version: %v", err)
			}

			newTag := opts.P.Version(ver.String())
			if opts.Prefix != "" {
				newTag = opts.Prefix + newTag
			}
			opts.Result.Tag.New = newTag

			// We don't need previous tag here; messages are simpler.
			opts.P.Printf("%s set tag %s\n", opts.P.Symbols.Bullet, newTag)

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
	return cmd
}
