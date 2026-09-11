package cli

import (
	"fmt"
	"time"

	"github.com/floppy-notes/floppy/internal/database"
	"github.com/floppy-notes/floppy/internal/index"
	"github.com/floppy-notes/floppy/internal/item"
	"github.com/spf13/cobra"
)

type updateFlags struct {
	ID       string
	Title    string
	Replace  bool
	Body     string
	Tags     []string
	Related  []string
	Status   string
	Due      string
	RemindAt string
	Quiet    bool
}

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing item",
	}

	cmd.AddCommand(
		newUpdateNoteCmd(),
		newUpdateTaskCmd(),
		newUpdateReminderCmd(),
	)

	return cmd
}

func addCommonUpdateFlags(cmd *cobra.Command, flags *updateFlags) {
	cmd.Flags().BoolVarP(&flags.Replace, "replace", "r", false, "should replace content")
	cmd.Flags().StringVarP(&flags.ID, "id", "i", "", "item's id")
	cmd.Flags().StringVarP(&flags.Title, "title", "t", "", "set new item's title")
	cmd.Flags().StringVarP(&flags.Body, "body", "b", "", "set new item's body")
	cmd.Flags().StringSliceVar(&flags.Tags, "tags", nil, "set new item's tags")
	cmd.Flags().StringSliceVar(&flags.Related, "related", nil, "set new item's related")
	cmd.Flags().BoolVarP(&flags.Quiet, "quiet", "q", false, "suppress output")
}

func newUpdateNoteCmd() *cobra.Command {
	flags := new(updateFlags)

	cmd := &cobra.Command{
		Use:   "note",
		Short: "Update note",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := database.Open(database.WithPath(vaultPath))

			if err != nil {
				return fmt.Errorf("opening database: %w", err)
			}

			defer db.Conn.Close()

			index.Update(db, index.UpdateRequest{
				ID:      flags.ID,
				Type:    item.TypeNote,
				Title:   flags.Title,
				Body:    flags.Body,
				Tags:    flags.Tags,
				Related: flags.Related,
			}, flags.Replace)

			return nil
		},
	}
	addCommonUpdateFlags(cmd, flags)
	return cmd

}

func newUpdateTaskCmd() *cobra.Command {
	flags := new(updateFlags)

	cmd := &cobra.Command{
		Use:   "task",
		Short: "Update task",
		RunE: func(cmd *cobra.Command, args []string) error {
			var dueStr string

			if flags.Due != "" {
				due, err := parseDate(flags.Due)
				if err != nil {
					return fmt.Errorf("invalid --due: %w", err)
				}
				dueStr = due.Format(time.DateOnly)
			}

			db, err := database.Open(database.WithPath(vaultPath))

			if err != nil {
				return fmt.Errorf("opening database: %w", err)
			}

			defer db.Conn.Close()

			index.Update(db, index.UpdateRequest{
				ID:      flags.ID,
				Type:    item.TypeTask,
				Title:   flags.Title,
				Body:    flags.Body,
				Tags:    flags.Tags,
				Related: flags.Related,
				Due:     dueStr,
				Status:  flags.Status,
			}, flags.Replace)

			return nil
		},
	}

	addCommonUpdateFlags(cmd, flags)
	cmd.Flags().StringVarP(&flags.Due, "due", "d", "", "due date (YYYY-MM-DD)")
	cmd.Flags().StringVarP(&flags.Status, "status", "s", "", "task status (open, done, archived)")
	return cmd

}

func newUpdateReminderCmd() *cobra.Command {
	flags := new(updateFlags)

	cmd := &cobra.Command{
		Use:   "reminder",
		Short: "Update reminder",
		RunE: func(cmd *cobra.Command, args []string) error {
			var remindAtStr string

			if flags.RemindAt != "" {
				remindAt, err := parseDate(flags.RemindAt)
				if err != nil {
					return fmt.Errorf("invalid --remind-at: %w", err)
				}
				remindAtStr = remindAt.Format(time.RFC3339)
			}

			db, err := database.Open(database.WithPath(vaultPath))

			if err != nil {
				return fmt.Errorf("opening database: %w", err)
			}

			defer db.Conn.Close()

			index.Update(db, index.UpdateRequest{
				ID:       flags.ID,
				Type:     item.TypeReminder,
				Title:    flags.Title,
				Body:     flags.Body,
				Tags:     flags.Tags,
				Related:  flags.Related,
				RemindAt: remindAtStr,
				Status:   flags.Status,
			}, flags.Replace)

			return nil
		},
	}

	addCommonUpdateFlags(cmd, flags)
	cmd.Flags().StringVar(&flags.RemindAt, "remind-at", "", "when to be reminded (YYYY-MM-DD or YYYY-MM-DD HH:MM)")
	cmd.Flags().StringVarP(&flags.Status, "status", "s", "", "reminder status (pending, fired, dismissed)")
	return cmd

}

func init() {
	rootCmd.AddCommand(newUpdateCmd())
}
