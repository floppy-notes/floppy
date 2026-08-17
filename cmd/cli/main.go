package main

import (
	"fmt"
	"os"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/item"
)

func main() {

	db, err := database.Open(
		database.WithPath("testdata"),
	)

	db.Conn.Ping()

	if err != nil {
		fmt.Println(err)
	}

	rawData, err := os.ReadFile("testdata/floppy_vault/notes/2026/08/2026-08-12-0001.example.md")

	if err != nil {
		fmt.Println(err)
	}
	parsedItem, err := item.Parse(rawData)

	if err != nil {
		fmt.Println(err)
	}
	err = parsedItem.Front.Validate()

	if err != nil {
		fmt.Println(err)
	}

}
