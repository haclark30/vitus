package cmd

import (
	"database/sql"
	"log"
	"time"

	tslc "github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
)

type HeartChart struct {
	tslc.Model
	db *sql.DB
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

func NewHeartChart(db *sql.DB, width, height int) HeartChart {
	dataSet := GetHeartData(db, 0, 1)
	chart := tslc.New(
		width,
		height,
		tslc.WithDataSetTimeSeries("heart data", dataSet),
		tslc.WithXLabelFormatter(tslc.HourTimeLabelFormatter()),
		tslc.WithYRange(50, 175),
	)
	return HeartChart{
		Model: chart,
		db:    db,
	}
}
