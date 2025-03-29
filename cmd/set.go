package cmd

import "github.com/spf13/cobra"

func CreateSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "set",
	}

	return cmd
}
