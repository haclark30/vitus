package main

import (
	"log"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/haclark30/vitus/cmd"
	"github.com/stackus/dotenv"
)

const fitbitUrl = "https://api.fitbit.com/"

func main() {
	err := dotenv.Load()
	if os.Getenv("LOGLEVEL") == "DEBUG" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	if err != nil {
		log.Fatal(err)
	}
	f, _ := tea.LogToFile("debug.log", "debug")
	defer f.Close()
	cmd.Execute()
}
