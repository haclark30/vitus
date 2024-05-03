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

type activeState int

const (
	stepsActive activeState = iota
	weightActive
	numStates // used to keep track of number of states
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
	activeState activeState
}

func (m model) incrementState() activeState {
	return (m.activeState + 1) % numStates
}

func (m model) decrementState() activeState {
	return (m.activeState - 1 + numStates) % numStates
}

func (m model) Init() tea.Cmd {
	m.stepsChart.Draw()
	m.weightChart.DrawXYAxisAndLabel()
	m.weightChart.DrawBrailleAll()
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	forwardmsg := false
	activeChange := false
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
		case "tab":
			m.activeState = m.incrementState()
			activeChange = true
		case "shift+tab":
			m.activeState = m.decrementState()
			activeChange = true
		}
	}
	if activeChange {
		m.weightChart.Blur()
		m.stepsChart.Canvas.Blur()
		switch m.activeState {
		case stepsActive:
			m.stepsChart.Canvas.Focus()
		case weightActive:
			m.weightChart.Focus()
		}
	}
	if forwardmsg {
		switch m.activeState {
		case weightActive:
			m.weightChart.Model, _ = m.weightChart.Model.Update(msg)
			m.weightChart.DrawBrailleAll()
		case stepsActive:
			m.stepsChart, _ = m.stepsChart.Update(msg)
			m.stepsChart.Draw()
		}
	}
	return m, nil
}

func (m model) View() string {
	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	var chartView string

	switch m.activeState {
	case stepsActive:
		chartView = defaultStyle.Render(m.stepsChart.View())
	case weightActive:
		chartView = defaultStyle.Render(m.weightChart.View())
	}

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
	m := model{db, stepsChart, weightChart, zoneManager, stepsActive}
	if _, err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		log.Fatal(err)
	}
}
