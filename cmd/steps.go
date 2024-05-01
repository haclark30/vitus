package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/NimbleMarkets/ntcharts/barchart"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

type StepsChart struct {
	barchart.Model
	db           *sql.DB
	stepsData    []barchart.BarData
	activeIdx    int
	startDayDiff int
	endDayDiff   int
	zoneManager  *zone.Manager
}

func (s StepsChart) Update(msg tea.Msg) (StepsChart, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			// TODO: handle index out of range
			if s.Canvas.Focused() {
				s.Clear()
				s.startDayDiff--
				s.endDayDiff--
				s.stepsData = GetStepsData(s.db, s.startDayDiff, s.endDayDiff)
				s.PushAll(s.stepsData)
			}
		case "up", "k":
			// TODO: handle index out of range
			if s.Canvas.Focused() {
				s.Clear()
				s.startDayDiff++
				s.endDayDiff++
				s.stepsData = GetStepsData(s.db, s.startDayDiff, s.endDayDiff)
				s.PushAll(s.stepsData)
			}
		case "right", "l":
			// go forward through bars
			if s.Canvas.Focused() {
				s.stepsData[s.activeIdx].Values[0].Style = s.stepsData[s.activeIdx].Values[0].Style.Faint(false)
				if s.activeIdx < len(s.stepsData)-1 {
					s.activeIdx++
				} else {
					s.activeIdx = 0
				}
				s.stepsData[s.activeIdx].Values[0].Style = s.stepsData[s.activeIdx].Values[0].Style.Faint(true)
				s.Draw()
			}
		case "left", "h":
			// go backward through bars
			if s.Canvas.Focused() {
				s.stepsData[s.activeIdx].Values[0].Style = s.stepsData[s.activeIdx].Values[0].Style.Faint(false)
				if s.activeIdx > 0 {
					s.activeIdx--
				} else {
					s.activeIdx = len(s.stepsData) - 1
				}
				s.stepsData[s.activeIdx].Values[0].Style = s.stepsData[s.activeIdx].Values[0].Style.Faint(true)
				s.Draw()
			}
		}
	}
	return s, nil
}

func (s StepsChart) View() string {
	currentDay := time.Now().Add(time.Duration(24*s.startDayDiff) * time.Hour)
	currentDayStr := currentDay.Format("2006-01-02")

	currentHourSteps := s.stepsData[s.activeIdx].Values[0].Value
	return lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Align(lipgloss.Center).Render("steps per hour\n"+currentDayStr+"\n"+s.Model.View()),
		lipgloss.NewStyle().Width(10).Render(fmt.Sprintf("%d steps", int(currentHourSteps))),
	)
}

func NewStepsChart(db *sql.DB, width, height int, zoneManager *zone.Manager) StepsChart {
	stepsData := GetStepsData(db, 0, 1)
	barChart := barchart.New(width, height, barchart.WithDataSet(stepsData))

	var activeIdx int
	for idx, stepCol := range stepsData {
		if stepCol.Values[0].Value > 0 {
			stepCol.Values[0].Style = stepCol.Values[0].Style.Faint(true)
			activeIdx = idx
			break
		}
	}

	barChart.SetZoneManager(zoneManager)
	return StepsChart{
		Model:        barChart,
		db:           db,
		stepsData:    stepsData,
		activeIdx:    activeIdx,
		startDayDiff: 0,
		endDayDiff:   1,
		zoneManager:  zoneManager,
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
