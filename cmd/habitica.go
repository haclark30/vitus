package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/haclark30/vitus/habitica"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(habiticaCmd)
}

var habiticaCmd = &cobra.Command{
	Use:   "habitica",
	Short: "habitica stats",
	Run:   habiticaRun,
}

func habiticaRun(cmd *cobra.Command, args []string) {

	habClient := habitica.NewHabiticaClient(os.Getenv("HABITICA_USER_ID"), os.Getenv("HABITICA_API_KEY"))
	resp, _ := habClient.Get("https://habitica.com/api/v3/tasks/user")
	defer resp.Body.Close()
	h, _ := io.ReadAll(resp.Body)
	fmt.Println(string(h))
}
