package cmd

import (
	"database/sql"
	"log"
	"time"

	tslc "github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

type WeightChart struct {
	tslc.Model
	zoneManager *zone.Manager
}

func (w WeightChart) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Align(lipgloss.Center).Render("weight per day\n"+w.Model.View()),
	)
}

func NewWeightChart(db *sql.DB, width, height int, zoneManager *zone.Manager) WeightChart {
	weightChart := tslc.New(width, height, tslc.WithZoneManager(zoneManager))

	now := time.Now()
	weightChart = LoadWeightChart(db, weightChart, time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()))
	return WeightChart{weightChart, zoneManager}
}
func LoadWeightChart(db *sql.DB, chart tslc.Model, startDate time.Time) tslc.Model {
	stmt, err := db.Prepare(
		`SELECT date(date), weight FROM WeightRecords where date >= ?`,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(startDate.Format("2006-01-02"))
	if err != nil {
		log.Fatal()
	}
	defer rows.Close()

	var weightRecords []float64
	var weightTimes []string

	for rows.Next() {
		var wr float64
		var wt string
		err = rows.Scan(&wt, &wr)
		if err != nil {
			log.Fatal(err)
		}
		weightRecords = append(weightRecords, wr)
		weightTimes = append(weightTimes, wt)
	}

	minWeight, maxWeight := weightRecords[0], weightRecords[0]
	for i := range weightTimes {
		wt, err := time.Parse("2006-01-02", weightTimes[i])
		if err != nil {
			log.Fatal("error parsing date from db", err)
		}
		chart.Push(tslc.TimePoint{Time: wt, Value: weightRecords[i]})
		if weightRecords[i] < minWeight {
			minWeight = weightRecords[i]
		}

		if weightRecords[i] > maxWeight {
			maxWeight = weightRecords[i]
		}
	}

	chart.SetYRange(150, 170)
	chart.SetViewYRange(150, 170)
	chart.XLabelFormatter = tslc.DateTimeLabelFormatter()
	xEndTime, err := time.Parse("2006-01-02", weightTimes[len(weightTimes)-1])
	if err != nil {
		log.Fatal(err)
	}
	chart.SetViewTimeRange(startDate, xEndTime)

	return chart
}
