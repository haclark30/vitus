package main

import (
	"github.com/haclark30/vitus/cmd"
	"github.com/stackus/dotenv"
)

const fitbitUrl = "https://api.fitbit.com/"

func main() {
	dotenv.Load()
	cmd.Execute()
	//
	// habClient := habitica.NewHabiticaClient(os.Getenv("HABITICA_USER_ID"), os.Getenv("HABITICA_API_KEY"))
	// resp, _ = habClient.Get("https://habitica.com/api/v3/tasks/user")
	// defer resp.Body.Close()
	// h, _ := io.ReadAll(resp.Body)
	// fmt.Println(string(h))
}
