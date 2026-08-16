package main

import (
	"fmt"
	"os"

	"github.com/floppy-notes/floppy/internal/item"
)

func main() {
	rawData, err := os.ReadFile("testdata/floppy_vault/notes/2026/08/2026-08-12-0001-example.md")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(item.Parse(rawData))
}
