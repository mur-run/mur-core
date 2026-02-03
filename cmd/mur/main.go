package main

import (
	"os"

	"github.com/karajanchang/murmur-ai/cmd/mur/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
