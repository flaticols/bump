package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/flaticols/bump/internal"
	"github.com/spf13/cobra"
)

func CreateUndoCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Remove the latest semver git tag",
		Long:  "Remove the latest semver git tag both locally and from the remote repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			allowNoRemotes, _ := cmd.Root().PersistentFlags().GetBool("no-remotes")

			// Get current version (latest tag)
			ver, err := opts.GitDetailer.GetCurrentVersion()
			var tagErr internal.SemVerTagError
			if err != nil {
				if errors.As(err, &tagErr) {
					if tagErr.NoTags {
						fmt.Println(opts.ErrPrinter("no tags found to remove"))
						os.Exit(1)
					}
					fmt.Println(opts.ErrPrinter("tag '%s' is not a valid semver tag", tagErr.Tag))
					os.Exit(1)
				}
				return err
			}

			tag := fmt.Sprintf("v%s", ver.String())
			fmt.Printf("removing tag %s\n", opts.InfoPrinter(tag))

			// Delete the tag locally
			if err := removeLocalGitTag(tag); err != nil {
				return err
			}
			fmt.Println(opts.OkPrinter("local tag removed"))

			// If remotes are available, delete the tag from remote as well
			if !allowNoRemotes {
				if err := removeRemoteGitTag(tag); err != nil {
					fmt.Println(opts.WarningPrinter("failed to remove remote tag: %v", err))
					fmt.Println(opts.WarningPrinter("local tag was removed, but remote tag may still exist"))
					return err
				}
				fmt.Println(opts.OkPrinter("remote tag removed"))
			}

			return nil
		},
	}

	return cmd
}

// removeLocalGitTag deletes a git tag locally
func removeLocalGitTag(tag string) error {
	cmd := exec.Command("git", "tag", "-d", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error removing local git tag: %v - %s", err, string(output))
	}
	return nil
}

// removeRemoteGitTag deletes a git tag from the remote repository
func removeRemoteGitTag(tag string) error {
	cmd := exec.Command("git", "push", "--delete", "origin", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error removing remote git tag: %v - %s", err, string(output))
	}
	return nil
}
