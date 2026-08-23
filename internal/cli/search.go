package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchLimit int

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search full-text",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(args)
		return nil
	},
}

func init() {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 20, "result limit")
	rootCmd.AddCommand(searchCmd)
}
