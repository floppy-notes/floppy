package cli

import (
	"fmt"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/index"
	"github.com/spf13/cobra"
)

var indexRebuild bool

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index or rebuild vault",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !indexRebuild {
			return fmt.Errorf("incremental indexing not implemented yet, use --rebuild")
		}

		db, err := database.Open(database.WithPath(vaultPath))
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer db.Conn.Close()

		if err := index.Rebuild(db, vaultPath); err != nil {
			return fmt.Errorf("rebuilding index: %w", err)
		}

		fmt.Println("index rebuilt for vault:", vaultPath)
		return nil
	},
}

func init() {
	indexCmd.Flags().BoolVarP(&indexRebuild, "rebuild", "r", false, "rebuild vault's index")
	rootCmd.AddCommand(indexCmd)
}
