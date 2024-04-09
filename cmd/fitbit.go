package cmd

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/guptarohit/asciigraph"
	"github.com/haclark30/vitus/db"
	"github.com/haclark30/vitus/fitbit"
	"github.com/spf13/cobra"
)

const fitbitUrl = "https://api.fitbit.com/"

var fitbitCmd = &cobra.Command{
	Use:   "fitbit",
	Short: "fitbit stats",
	Run:   fitbitRun,
}

var fitbitLoadCmd = &cobra.Command{
	Use:   "load",
	Short: "load fitbit data to db",
	Run:   fitbitLoad,
}

var fitbitApiCmd = &cobra.Command{
	Use:   "api",
	Short: "call api with given url",
	Run:   fitbitApi,
}

var fitbitWaterCmd = &cobra.Command{
	Use:   "water",
	Short: "add water",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run:   fitbitWater,
}

func init() {
	fitbitCmd.AddCommand(fitbitLoadCmd)
	fitbitCmd.AddCommand(fitbitApiCmd)
	fitbitCmd.AddCommand(fitbitWaterCmd)
}

func fitbitRun(cmd *cobra.Command, args []string) {
	client := fitbit.NewFitbitClient()
	heartRate := fitbit.GetHeartDay(client)
	stepsDay := fitbit.GetStepsDay(client)
	heartData := make([]float64, 0)
	stepsData := make([]float64, 0)
	heartIndex := 0
	for _, step := range stepsDay.ActivitiesStepsIntra.Dataset {
		var heart fitbit.IntradayData
		if heartIndex < len(heartRate.ActivitiesHeartIntraday.Dataset) {
			heart = heartRate.ActivitiesHeartIntraday.Dataset[heartIndex]
		}
		if step.Time == heart.Time {
			heartData = append(heartData, float64(heart.Value))
			heartIndex++
		} else {
			heartData = append(heartData, 0)
		}
		stepsData = append(stepsData, float64(step.Value))
	}
	graph := asciigraph.Plot(
		heartData,
		asciigraph.Height(10),
		asciigraph.Width(100))
	fmt.Println(graph)
	fmt.Println()

	graph = asciigraph.Plot(
		stepsData,
		asciigraph.Height(10),
		asciigraph.Width(100))
	fmt.Println(graph)
	fmt.Println()

	var weightData []float64
	db := db.GetDb()
	rows, err := db.Query("SELECT weight FROM WeightRecords WHERE date >= date('2024-01-01') ORDER BY date")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var weight float64
		if err := rows.Scan(&weight); err != nil {
			log.Fatal(err)
		}
		weightData = append(weightData, weight)
	}

	graph = asciigraph.Plot(
		weightData,
		asciigraph.Height(10),
	)
	fmt.Println(graph)
	sleepData := fitbit.GetSleepToday(client)
	sleepHours := sleepData.Summary.TotalMinutesAsleep / 60
	sleepMinutes := sleepData.Summary.TotalMinutesAsleep % 60

	fmt.Printf("slept for %d hours %d minutes\n", sleepHours, sleepMinutes)

	activityToday := fitbit.GetActivitiesToday(client)
	fmt.Printf("steps today: %d\n", activityToday.Summary.Steps)
	fmt.Printf("active minutes today: %d\n", activityToday.Summary.VeryActiveMinutes)
}

func fitbitLoad(cmd *cobra.Command, args []string) {
	client := fitbit.NewFitbitClient()
	db := db.GetDb()
	loadFitbitDb(client, db)
}

func fitbitApi(cmd *cobra.Command, args []string) {
	client := fitbit.NewFitbitClient()
	resp, err := client.Get(fmt.Sprintf("%s/%s", fitbitUrl, args[0]))
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}

func fitbitWater(cmd *cobra.Command, args []string) {
	client := fitbit.NewFitbitClient()
	oz, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal(err)
	}
	fitbit.AddWater(client, oz)
}

func loadFitbitDb(client *http.Client, db *sql.DB) error {
	currTime := time.Now()
	for currTime.Year() >= 2023 {
		weightData := fitbit.GetWeight(client, currTime, "30d")
		for _, w := range weightData.Weight {
			stmt, err := db.Prepare("INSERT INTO WeightRecords (date, weight) VALUES (?, ?)")
			if err != nil {
				log.Fatal(err)
			}
			defer stmt.Close()

			_, err = stmt.Exec(w.Date, w.Weight)
			if err != nil {
				log.Fatal(err)
			}
		}
		currTime = currTime.Add(-30 * 24 * time.Hour)
	}
	return nil
}
