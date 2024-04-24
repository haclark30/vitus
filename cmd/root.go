package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vitus",
	Short: "my cli for things in my life",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("test")
	},
}

func init() {
	rootCmd.AddCommand(fitbitCmd)
	rootCmd.AddCommand(ynabCmd)
	rootCmd.AddCommand(createDbCmd)
	rootCmd.AddCommand(teaCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
