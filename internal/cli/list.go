package cli

import (
	"fmt"
	"time"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/index"
	"github.com/floppy-notes/floppy/internal/item"
	"github.com/spf13/cobra"
)

type ListFlags struct {
	Type   string
	Status string
	Tag    string
	Since  string
	Until  string
	Limit  int
}

func (lf *ListFlags) validate() error {
	if lf.Type != "" && !item.IsValidItemType(lf.Type) {
		return fmt.Errorf("invalid item type: %s", lf.Type)
	}

	if lf.Status != "" && (!item.IsValidTaskStatus(lf.Status) && !item.IsValidReminderStatus(lf.Status)) {
		return fmt.Errorf("non existents status: %s", lf.Status)
	}

	if lf.Since != "" {
		t, err := parseDate(lf.Since)
		if err != nil {
			return fmt.Errorf("invalid since date: %w", err)
		}
		lf.Since = t.Format(time.RFC3339)
	}

	if lf.Until != "" {
		t, err := parseDate(lf.Until)
		if err != nil {
			return fmt.Errorf("invalid until date: %w", err)
		}
		lf.Until = t.Format(time.RFC3339)
	}

	return nil
}

var listFlags ListFlags

var listCmd = &cobra.Command{
	Use:   "list <params>",
	Short: "list items",
	RunE: func(cmd *cobra.Command, args []string) error {

		if err := listFlags.validate(); err != nil {
			return fmt.Errorf("validating flags: %w", err)
		}

		db, err := database.Open(database.WithPath(vaultPath))
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer db.Conn.Close()

		results, err := index.List(db, database.ListFilter{
			Type:   listFlags.Type,
			Status: listFlags.Status,
			Tag:    listFlags.Tag,
			Since:  listFlags.Since,
			Until:  listFlags.Until,
		}, listFlags.Limit)

		fmt.Println(results)
		return nil
	},
}

func init() {
	listCmd.Flags().StringVarP(&listFlags.Type, "type", "t", "", "item type")
	listCmd.Flags().StringVarP(&listFlags.Status, "status", "s", "", "item status")
	listCmd.Flags().StringVar(&listFlags.Tag, "tag", "", "item tags")
	listCmd.Flags().StringVar(&listFlags.Since, "since", "", "item created since")
	listCmd.Flags().StringVar(&listFlags.Until, "until", "", "item created until")
	listCmd.Flags().IntVarP(&listFlags.Limit, "limit", "l", 10, "list limit")

	rootCmd.AddCommand(listCmd)
}
