package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchLimit int

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search full-text",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(args)
		fmt.Println(cmd.Flags().GetInt("limit"))
		return nil
	},
}

func init() {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 20, "result limit")
	rootCmd.AddCommand(searchCmd)
}
