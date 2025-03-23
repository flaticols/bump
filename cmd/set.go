package cmd

import "github.com/spf13/cobra"

func CreateSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [version]",
		Short: "set [semver]",
		Long:  "set [semver] set the version of the project",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	return cmd
}
