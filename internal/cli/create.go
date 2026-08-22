package cli

import (
	"fmt"
	"time"

	"github.com/floppy-notes/floppy/internal/item"
	"github.com/spf13/cobra"
)

type createFlags struct {
	Tags     []string
	Body     string
	Related  []string
	Due      string
	RemindAt string
	Quiet    bool
}

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new item",
	}
	cmd.AddCommand(
		newCreateNoteCmd(),
		newCreateTaskCmd(),
		newCreateReminderCmd(),
	)
	return cmd
}

func addCommonFlags(cmd *cobra.Command, flags *createFlags) {
	cmd.Flags().StringSliceVarP(&flags.Tags, "tags", "t", nil, "tags to attach to the item")
	cmd.Flags().StringSliceVarP(&flags.Related, "related", "r", nil, "IDs of related items")
	cmd.Flags().StringVarP(&flags.Body, "body", "b", "", "item body content")
	cmd.Flags().BoolVarP(&flags.Quiet, "quier", "q", false, "suppress output")
}

func newCreateNoteCmd() *cobra.Command {
	flags := new(createFlags)
	cmd := &cobra.Command{
		Use:   "note",
		Short: "Create a new note",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("note", item.TypeNote, flags)
			return nil
		},
	}
	addCommonFlags(cmd, flags)
	return cmd
}

func newCreateTaskCmd() *cobra.Command {
	flags := new(createFlags)
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Create a new task",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			due, err := parseDate(flags.Due)
			if err != nil {
				return fmt.Errorf("invalid --due: %w", err)
			}
			fmt.Println("task", item.TypeTask, flags, due)
			return nil
		},
	}
	addCommonFlags(cmd, flags)
	cmd.Flags().StringVarP(&flags.Due, "due", "d", today(), "due date (YYYY-MM-DD)")
	return cmd
}

func newCreateReminderCmd() *cobra.Command {
	flags := new(createFlags)
	cmd := &cobra.Command{
		Use:   "reminder",
		Short: "Create a new reminder",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			remindAt, err := parseDate(flags.RemindAt)
			if err != nil {
				return fmt.Errorf("invalid --remind-at: %w", err)
			}
			fmt.Println("reminder", item.TypeReminder, flags, remindAt)
			return nil
		},
	}
	addCommonFlags(cmd, flags)
	cmd.Flags().StringVar(&flags.RemindAt, "remind-at", nowMinute(), "when to be reminded (YYYY-MM-DD or YYYY-MM-DD HH:MM)")
	return cmd
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	layouts := []string{"2006-01-02T15:04", "2006-01-02 15:04", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("%q: expected YYYY-MM-DD or YYYY-MM-DD HH:MM", s)
}

func today() string {
	return time.Now().Format("2006-01-02")
}

func nowMinute() string {
	return time.Now().Format("2006-01-02 15:04")
}

func init() {
	rootCmd.AddCommand(newCreateCmd())
}
