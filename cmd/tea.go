package cmd

import (
	"database/sql"
	"log"
	"os"

	"github.com/haclark30/vitus/db"
	zone "github.com/lrstanley/bubblezone"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var teaCmd = &cobra.Command{
	Use: "tea",
	Run: runTea,
}

var defaultStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")).
	Align(lipgloss.Center)

type model struct {
	db          *sql.DB
	stepsChart  StepsChart
	weightChart WeightChart
	zoneManager *zone.Manager
}

func (m model) Init() tea.Cmd {
	m.stepsChart.Draw()
	m.weightChart.DrawXYAxisAndLabel()
	m.weightChart.DrawBrailleAll()
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	forwardmsg := false
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "pgup", "pgdn":
			forwardmsg = true
		case "down", "j":
			forwardmsg = true
		case "up", "k":
			forwardmsg = true
		case "q":
			return m, tea.Quit
		case "right", "l":
			forwardmsg = true
		case "left", "h":
			forwardmsg = true
		}
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			m.weightChart.Blur()
			m.stepsChart.Canvas.Blur()

			switch {
			case m.zoneManager.Get(m.weightChart.ZoneID()).InBounds(msg):
				m.weightChart.Focus()
			case m.zoneManager.Get(m.stepsChart.ZoneID()).InBounds(msg):
				m.stepsChart.Canvas.Focus()
			}
		}
		forwardmsg = true
	}
	if forwardmsg {
		switch {
		case m.weightChart.Focused():
			m.weightChart.Model, _ = m.weightChart.Model.Update(msg)
			m.weightChart.DrawBrailleAll()
		case m.stepsChart.Canvas.Focused():
			m.stepsChart, _ = m.stepsChart.Update(msg)
			m.stepsChart.Draw()
		}
	}
	return m, nil
}

func (m model) View() string {
	// call bubblezone Manager.Scan() at root model
	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	chartView := lipgloss.JoinVertical(lipgloss.Center,
		defaultStyle.Render(m.stepsChart.View()),
		defaultStyle.Render(m.weightChart.View()),
	)

	chartView = lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(chartView)

	return m.zoneManager.Scan(chartView)
}

func runTea(cmd *cobra.Command, args []string) {
	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	height := 16

	db := db.GetDb()
	zoneManager := zone.New()

	weightChart := NewWeightChart(db, width-10, height, zoneManager)
	stepsChart := NewStepsChart(db, width-20, height, zoneManager)

	stepsChart.Canvas.Focus()
	m := model{db, stepsChart, weightChart, zoneManager}
	if _, err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		log.Fatal(err)
	}
}
