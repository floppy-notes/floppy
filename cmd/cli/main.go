package main

import (
	"fmt"

	"github.com/floppy-notes/floppy/internal/vault"
)

func main() {
	fmt.Println(vault.Walk("testdata"))

	// db, err := database.Open(
	// 	database.WithPath("testdata"),
	// )

	// if err != nil {
	// 	fmt.Println(err)
	// }

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
