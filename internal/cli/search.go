package cli

import (
	"fmt"
	"strings"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/index"
	"github.com/spf13/cobra"
)

var searchLimit int

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search full-text",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.Open(database.WithPath(vaultPath))
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer db.Conn.Close()

		query := strings.Join(args, " ")
		results, err := index.Search(db, query, searchLimit)
		if err != nil {
			return fmt.Errorf("searching: %w", err)
		}

		return writeItems(cmd.OutOrStdout(), results)
	},
}

func init() {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 20, "result limit")
	rootCmd.AddCommand(searchCmd)
}
