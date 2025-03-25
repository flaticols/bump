package internal

import (
	"os"
	"path"
)

// SetBumpWd changes the current working directory to the specified directory and returns an error if the operation fails.
func SetBumpWd(repoDirectory string) error {
	if repoDirectory == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		repoDirectory = wd
	}

	err := os.Chdir(path.Clean(repoDirectory))
	if err != nil {
		return err
	}

	return nil
}
