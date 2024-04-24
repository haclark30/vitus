package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/haclark30/vitus/db"
	zone "github.com/lrstanley/bubblezone"
	"github.com/spf13/cobra"

	"github.com/NimbleMarkets/ntcharts/barchart"
	tslc "github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var teaCmd = &cobra.Command{
	Use: "tea",
	Run: runTea,
}

type model struct {
	db           *sql.DB
	stepsChart   barchart.Model
	stepsData    []barchart.BarData
	activeIdx    int
	weightChart  WeightChart
	startDayDiff int
	endDayDiff   int
	zoneManager  *zone.Manager
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
		// commands for hjkl?
		case "up", "down", "left", "right", "pgup", "pgdown":
			m.stepsChart.Clear()
			m.startDayDiff--
			m.endDayDiff--
			m.stepsData = GetStepsData(m.db, m.startDayDiff, m.endDayDiff)
			m.stepsChart.PushAll(m.stepsData)
			forwardmsg = true
		case "q":
			return m, tea.Quit
			// add commands for home/end?
		case "tab":
			// go forward through bars
			if m.stepsChart.Canvas.Focused() {
				m.stepsData[m.activeIdx].Values[0].Style = m.stepsData[m.activeIdx].Values[0].Style.Faint(false)
				if m.activeIdx < len(m.stepsData)-1 {
					m.activeIdx++
				} else {
					m.activeIdx = 0
				}
				m.stepsData[m.activeIdx].Values[0].Style = m.stepsData[m.activeIdx].Values[0].Style.Faint(true)
				m.stepsChart.Draw()
			}
		case "shift+tab":
			// go backward through bars
			if m.stepsChart.Canvas.Focused() {
				m.stepsData[m.activeIdx].Values[0].Style = m.stepsData[m.activeIdx].Values[0].Style.Faint(false)
				if m.activeIdx > 0 {
					m.activeIdx--
				} else {
					m.activeIdx = len(m.stepsData) - 1
				}
				m.stepsData[m.activeIdx].Values[0].Style = m.stepsData[m.activeIdx].Values[0].Style.Faint(true)
				m.stepsChart.Draw()
			}
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
	currentHourSteps := m.stepsData[m.activeIdx].Values[0].Value
	chartView := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63")). // purple
			Render(fmt.Sprintf("active %d\n", m.activeIdx)+"steps\n"+fmt.Sprintf("%d\n", int(currentHourSteps))+m.stepsChart.View()),
		m.weightChart.View(),
	)

	return m.zoneManager.Scan(chartView)
}

func runTea(cmd *cobra.Command, args []string) {
	width := 36 * 2
	height := 8 * 2

	db := db.GetDb()

	stepsData := GetStepsData(db, 0, 1)
	var activeIdx int
	for idx, stepCol := range stepsData {
		if stepCol.Values[0].Value > 0 {
			stepCol.Values[0].Style = stepCol.Values[0].Style.Faint(true)
			activeIdx = idx
			break
		}
	}
	now := time.Now()
	zoneManager := zone.New()
	stepsChart := barchart.New(width, height, barchart.WithDataSet(stepsData))
	stepsChart.SetZoneManager(zoneManager)
	stepsChart.Canvas.Focus()

	weightChart := tslc.New(width, height,
		tslc.WithZoneManager(zoneManager))

	weightChart = LoadWeightChart(db, weightChart, time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()))

	m := model{db, stepsChart, stepsData, activeIdx, WeightChart{weightChart, zoneManager}, 0, 1, zoneManager}
	if _, err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		log.Fatal(err)
	}
}

func GetStepsData(db *sql.DB, startDayDiff, endDayDiff int) []barchart.BarData {

	stmt, err := db.Prepare(
		`SELECT
			strftime('%Y-%m-%d %H:00:00', datetime(time, 'unixepoch', 'localtime')) AS hour,
			SUM(steps) AS step_sum FROM StepsRecords
		WHERE hour >= date('now', ? || ' day') AND hour < date('now', ? || ' day')
		GROUP BY hour;`,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(startDayDiff, endDayDiff)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var steps []float64
	var stepTimes []string
	for rows.Next() {
		var step float64
		var stepTime string
		err = rows.Scan(&stepTime, &step)
		steps = append(steps, step)
		stepTimes = append(stepTimes, stepTime)
		if err != nil {
			log.Fatal(err)
		}
	}

	var stepsData []barchart.BarData
	for i := 0; i <= 23; i++ {
		if i < len(stepTimes) {
			stepTime, _ := time.Parse("2006-01-02 15:00:00", stepTimes[i])
			stepsData = append(stepsData, barchart.BarData{
				Label: stepTime.Format("15"),
				Values: []barchart.BarValue{
					{Name: "steps", Value: steps[i], Style: lipgloss.NewStyle().Foreground(lipgloss.Color("9"))},
				},
			})
		} else {
			// fill remaining hours with 0
			stepTime, _ := time.Parse("2006-01-02 15:00:00", stepTimes[len(stepTimes)-1])
			stepTime = stepTime.Add(time.Hour * time.Duration(i-len(stepTimes)+1))
			slog.Debug("adding stepTime", "time", stepTime)
			stepsData = append(stepsData, barchart.BarData{
				Label: stepTime.Format("15"),
				Values: []barchart.BarValue{
					{Name: "steps", Value: 0, Style: lipgloss.NewStyle().Foreground(lipgloss.Color("9"))},
				},
			})
		}
	}
	return stepsData
}
