package main

import (
	"os"

	"github.com/mur-run/mur-cli/cmd/mur/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
