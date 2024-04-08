package main

import (
	"github.com/haclark30/vitus/cmd"
	"github.com/stackus/dotenv"
)

const fitbitUrl = "https://api.fitbit.com/"

func main() {
	dotenv.Load()
	cmd.Execute()
}
