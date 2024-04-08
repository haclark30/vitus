package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/haclark30/vitus/ynab"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(ynabCmd)
}

var ynabCmd = &cobra.Command{
	Use:   "ynab",
	Short: "You Need a Budget stats",
	Run:   ynabRun,
}

func ynabRun(cmd *cobra.Command, args []string) {
	ynabClient := ynab.NewYnabClient(os.Getenv("YNAB_API_KEY"))
	resp, _ := ynabClient.Get("https://api.ynab.com/v1/budgets")
	defer resp.Body.Close()
	y, _ := io.ReadAll(resp.Body)
	fmt.Println(string(y))
}
