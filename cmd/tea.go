package cmd

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/haclark30/vitus/db"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type activeState int

const (
	stepsActive activeState = iota
	weightActive
	heartActive
	sleepActive
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
	heartChart  HeartChart
	activeState activeState
}

func (a activeState) String() string {
	switch a {
	case stepsActive:
		return "Steps"
	case weightActive:
		return "Weight"
	case heartActive:
		return "Heart"
	case sleepActive:
		return "Sleep"
	case numStates:
		return "None"
	default:
		return "Unknown State"
	}
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
	m.heartChart.DrawXYAxisAndLabel()
	m.heartChart.DrawBrailleAll()
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

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Copy().Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

func (m model) View() string {
	// TODO: make header
	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	doc := strings.Builder{}

	var renderedTabs []string

	// TODO: build border for any content

	// which tab is active
	switch m.activeState {
	case stepsActive:
		doc.WriteString(defaultStyle.Render(m.stepsChart.View()))
	case weightActive:
		doc.WriteString(defaultStyle.Render(m.weightChart.View()))
	case heartActive:
		doc.WriteString(defaultStyle.Render(m.heartChart.View()))
	}

	// build the tabs
	for state := stepsActive; state < numStates; state++ {
		var style lipgloss.Style
		style = style.BorderStyle(lipgloss.HiddenBorder())
		if state == m.activeState {
			style = style.BorderStyle(lipgloss.Border{Bottom: "_"}).BorderForeground(lipgloss.Color("63"))
		}
		renderedTabs = append(renderedTabs, style.Render(state.String()))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString("\n")
	doc.WriteString(row)
	doc.WriteString("\n")

	return docStyle.Width(width).Align(lipgloss.Center).Render(doc.String())
}

func runTea(cmd *cobra.Command, args []string) {
	width, height, _ := term.GetSize(int(os.Stdout.Fd()))

	db := db.GetDb()

	weightChart := NewWeightChart(db, width-10, height-10)
	stepsChart := NewStepsChart(db, width-20, height-10)
	heartChart := NewHeartChart(db, width-10, height-10)

	stepsChart.Canvas.Focus()
	m := model{db, stepsChart, weightChart, heartChart, stepsActive}
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}
}
