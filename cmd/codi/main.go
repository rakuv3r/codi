// Package main is the entry point for the codi CLI.
package main

import (
	"os"

	"github.com/rakuv3r/codi/internal/command"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
