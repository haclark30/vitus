package cmd

import (
	"database/sql"
	"log"
	"log/slog"
	"time"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	"github.com/NimbleMarkets/ntcharts/linechart"
	tslc "github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HeartChart struct {
	tslc.Model
	db           *sql.DB
	startDayDiff int
	endDayDiff   int
}

func GetHeartData(db *sql.DB, startDayDiff, endDayDiff int) []tslc.TimePoint {
	stmt, err := db.Prepare(
		`SELECT
			time, heartRate
		FROM HeartRateRecords
		WHERE datetime(time, 'unixepoch', 'localtime') >= date('now', ? || ' day') AND
			datetime(time, 'unixepoch', 'localtime') < date('now', ? || ' day');`,
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

	var timePts []tslc.TimePoint
	for rows.Next() {
		var timePt tslc.TimePoint
		var timeInt int64
		err = rows.Scan(&timeInt, &timePt.Value)
		if err != nil {
			log.Fatal(err)
		}
		timePt.Time = time.Unix(timeInt, 0)
		timePts = append(timePts, timePt)
	}
	return timePts
}

func (h HeartChart) Update(msg tea.Msg) (HeartChart, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		//TODO: combine repeated code
		case "down", "j":
			if h.Focused() {
				h.startDayDiff--
				h.endDayDiff--
				h.Clear()
				h.ClearAllData()
				newData := GetHeartData(h.db, h.startDayDiff, h.endDayDiff)
				slog.Debug("heart down",
					"startDay", h.startDayDiff,
					"endDay", h.endDayDiff,
					"readings", len(newData),
				)
				for _, t := range newData {
					h.PushDataSet("heart data", t)
				}
			}
		case "up", "k":
			if h.Focused() {
				h.startDayDiff++
				h.endDayDiff++
				h.Clear()
				h.ClearAllData()
				newData := GetHeartData(h.db, h.startDayDiff, h.endDayDiff)
				slog.Debug("heart up",
					"startDay", h.startDayDiff,
					"endDay", h.endDayDiff,
					"readings", len(newData),
				)
				for _, t := range newData {
					h.PushDataSet("heart data", t)
				}
			}
		}
		// TODO: clean this shit up
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		h.SetTimeRange(
			today.Add(time.Duration(h.startDayDiff*24*int(time.Hour))),
			today.Add(time.Duration(h.endDayDiff*24*int(time.Hour))),
		)
		h.SetViewTimeRange(
			today.Add(time.Duration(h.startDayDiff*24*int(time.Hour))),
			today.Add(time.Duration(h.endDayDiff*24*int(time.Hour))),
		)
	}
	return h, nil
}

func (h HeartChart) Draw() {
	h.DrawBrailleDataSets([]string{"heart data"})
	h.DrawLine(
		canvas.Float64Point{X: h.MinX(), Y: 114},
		canvas.Float64Point{X: h.MaxX(), Y: 114},
		runes.ArcLineStyle,
	)
	h.DrawLine(
		canvas.Float64Point{X: h.MinX(), Y: 139},
		canvas.Float64Point{X: h.MaxX(), Y: 139},
		runes.ArcLineStyle)
	h.DrawLine(
		canvas.Float64Point{X: h.MinX(), Y: 170},
		canvas.Float64Point{X: h.MaxX(), Y: 170},
		runes.ArcLineStyle)
}
func LocalHourLabelFormatter() linechart.LabelFormatter {
	return func(i int, v float64) string {
		t := time.Unix(int64(v), 0).Local()
		slog.Debug("label format", "time", t)
		return t.Format("15:04")
	}
}

func NewHeartChart(db *sql.DB, width, height, startDayDiff, endDayDiff int) HeartChart {
	dataSet := GetHeartData(db, startDayDiff, endDayDiff)
	chart := tslc.New(
		width,
		height,
		tslc.WithDataSetTimeSeries("heart data", dataSet),
		tslc.WithXLabelFormatter(LocalHourLabelFormatter()),
		tslc.WithYRange(50, 175),
	)
	chart.AutoMaxX = false
	chart.AutoMinX = false
	chart.SetStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("11")))

	return HeartChart{
		Model:        chart,
		db:           db,
		startDayDiff: 0,
		endDayDiff:   1,
	}
}
