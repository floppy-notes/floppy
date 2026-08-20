package main

import (
	"fmt"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/index"
)

const VAULT_ROOT_DIR = "testdata"

func main() {

	db, err := database.Open(
		database.WithPath(VAULT_ROOT_DIR),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	err = index.Rebuild(db, VAULT_ROOT_DIR)

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, t := range []string{"items", "items_fts", "tags", "edges"} {
		var n int
		db.Conn.QueryRow("SELECT COUNT(*) FROM " + t).Scan(&n)
		fmt.Printf("%s: %d\n", t, n)
	}

	// db.Conn.Ping()

	// rawData, err := os.ReadFile("testdata/floppy_vault/notes/2026/08/2026-08-12-0001.example.md")

	// if err != nil {
	// 	fmt.Println(err)
	// }
	// parsedItem, err := item.Parse(rawData)

	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err = parsedItem.Front.Validate()

	// if err != nil {
	// 	fmt.Println(err)
	// }

}
