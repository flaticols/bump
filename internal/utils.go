package internal

import (
	"os"
)

// setBumpWd changes the current working directory to the specified directory and returns an error if the operation fails.
func setBumpWd(wd string) error {
	if err := os.Chdir(wd); err != nil {
		return err
	}

	return nil
}
