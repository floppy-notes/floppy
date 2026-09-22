package cli

import (
	"fmt"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/index"
	"github.com/spf13/cobra"
)

type ShowFlags struct {
	BodyOnly bool
}

var showFlags ShowFlags

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show an item by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.Open(database.WithPath(vaultPath))

		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}

		defer db.Conn.Close()

		it, err := index.FindByID(db, args[0])

		if err != nil {
			return err
		}

		if showFlags.BodyOnly {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), it.Body)
			return err
		}

		return writeItem(cmd.OutOrStdout(), it)
	},
}

func init() {
	showCmd.Flags().BoolVarP(&showFlags.BodyOnly, "body-only", "b", false, "print only the item body")

	rootCmd.AddCommand(showCmd)
}
