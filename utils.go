package main

import (
	"os"
)

func setBumpWd(wd string) error {
	if err := os.Chdir(wd); err != nil {
		return err
	}

	return nil
}
