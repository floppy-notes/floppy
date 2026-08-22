package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var indexRebuild bool

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index or rebuild vault",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(args)
		fmt.Println(cmd.Flags().GetBool("rebuild"))
		return nil
	},
}

func init() {
	indexCmd.Flags().BoolVarP(&indexRebuild, "rebuild", "r", false, "rebuild vault's index")
	rootCmd.AddCommand(indexCmd)
}
