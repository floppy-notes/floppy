package main

import (
	"os"

	"github.com/floppy-notes/floppy/internal/cli"
)

func main() {

	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
