package cli

import (
	"github.com/spf13/cobra"
)

type updateFlags struct {
	ID       string
	Append   bool
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
	cmd.Flags().BoolVarP(&flags.Append, "append", "a", false, "should append content in the iten")
}

func newUpdateNoteCmd() *cobra.Command {
	flags := new(updateFlags)

	cmd := &cobra.Command{
		Use:   "note",
		Short: "Update note",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			return nil
		},
	}

	addCommonUpdateFlags(cmd, flags)

	return cmd

}

func newUpdateReminderCmd() *cobra.Command {
	flags := new(updateFlags)

	cmd := &cobra.Command{
		Use:   "reminder",
		Short: "Update reminder",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	addCommonUpdateFlags(cmd, flags)

	return cmd

}

func init() {
	rootCmd.AddCommand(newUpdateCmd())
}
