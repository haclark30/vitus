package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/haclark30/vitus/db"
	"github.com/haclark30/vitus/fitbit"
	"github.com/spf13/cobra"
)

var createDbCmd = &cobra.Command{
	Use:   "createdb",
	Short: "create a new sqlite database from scratch",
	Args:  cobra.ExactArgs(1),
	Run:   createDbRun,
}

func createDbRun(cmd *cobra.Command, args []string) {
	client := fitbit.NewFitbitClient()
	db := db.GetDb()
	var startDate time.Time
	switch dateArg := args[0]; dateArg {
	case "today":
		startDate = time.Now()
	case "yesterday":
		startDate = time.Now().Add(-24 * time.Hour)
	default:
		date, err := time.Parse("2006-01-02", dateArg)
		if err != nil {
			log.Fatalf("not a valid date: %s", dateArg)
		}
		startDate = date
	}
	loadFitbitDb(client, db, startDate)
}
func loadWeight(db *sql.DB, client *http.Client, startTime, endTime time.Time) {
	txn, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := txn.Prepare("INSERT OR REPLACE INTO WeightRecords (date, weight) VALUES (?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for endTime.Compare(startTime) >= 0 {
		slog.Debug("insert weight", "time", endTime)
		weightData := fitbit.GetWeight(client, endTime, "30d")
		for _, w := range weightData.Weight {
			slog.Debug("insert weight", "time", endTime)

			_, err = stmt.Exec(w.Date, w.Weight)
			if err != nil {
				log.Fatal(err)
			}
		}
		endTime = endTime.Add(-30 * 24 * time.Hour)
	}
	if err := txn.Commit(); err != nil {
		log.Fatal(err)
	}
	slog.Debug("done with weight")
}

func loadHeartRate(db *sql.DB, client *http.Client, startTime, endTime time.Time) {
	txn, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	for endTime.Compare(startTime) >= 0 {
		slog.Debug("insert heart", "time", endTime)
		hr := fitbit.GetHeartDay(client, endTime)
		for _, dayHr := range hr.ActivitiesHeartIntraday.Dataset {
			stmt, err := txn.Prepare("INSERT OR REPLACE INTO HeartRateRecords (time, heartRate) VALUES (?, ?)")
			if err != nil {
				log.Fatal(err)
			}
			defer stmt.Close()

			date := hr.ActivitiesHeart[0].DateTime
			tz, _ := time.Now().Zone()
			dateTime, err := time.Parse("2006-01-02T15:04:05 MST", fmt.Sprintf("%sT%s %s", date, dayHr.Time, tz))
			if err != nil {
				log.Fatal(err)
			}
			_, err = stmt.Exec(dateTime.Unix(), float64(dayHr.Value))
			if err != nil {
				log.Fatal(err)
			}
		}
		endTime = endTime.Add(-24 * time.Hour)
	}
	if err := txn.Commit(); err != nil {
		log.Fatal(err)
	}
	slog.Debug("done with heart")
}

func loadSteps(db *sql.DB, client *http.Client, startTime, endTime time.Time) {
	txn, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	for endTime.Compare(startTime) >= 0 {
		slog.Debug("insert steps", "time", endTime)
		stepData := fitbit.GetStepsDay(client, endTime)
		for _, steps := range stepData.ActivitiesStepsIntra.Dataset {
			stmt, err := txn.Prepare("INSERT OR REPLACE INTO StepsRecords (time, steps) VALUES (?, ?)")
			if err != nil {
				log.Fatal(err)
			}
			defer stmt.Close()

			date := stepData.ActivitiesSteps[0].DateTime
			tz, _ := time.Now().Zone()

			dateTime, err := time.Parse("2006-01-02T15:04:05 MST", fmt.Sprintf("%sT%s %s", date, steps.Time, tz))
			if err != nil {
				log.Fatal(err)
			}
			_, err = stmt.Exec(dateTime.Unix(), float64(steps.Value))
			if err != nil {
				log.Fatal(err)
			}
		}
		endTime = endTime.Add(-24 * time.Hour)
	}
	if err := txn.Commit(); err != nil {
		log.Fatal(err)
	}
	slog.Debug("done with steps")
}

func loadFitbitDb(client *http.Client, db *sql.DB, startTime time.Time) error {
	loadWeight(db, client, startTime, time.Now())
	loadHeartRate(db, client, startTime, time.Now())
	loadSteps(db, client, startTime, time.Now())
	return nil
}
