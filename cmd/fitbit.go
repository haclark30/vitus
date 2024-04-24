package cmd

import (
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
	"github.com/stackus/dotenv"
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

var client *http.Client

var loadVal int

func init() {
	dotenv.Load()
	client = fitbit.NewFitbitClient()
	fitbitLoadCmd.Flags().IntVar(&loadVal, "days", 30, "days")
	fitbitCmd.AddCommand(fitbitLoadCmd)
	fitbitCmd.AddCommand(fitbitApiCmd)
	fitbitCmd.AddCommand(fitbitWaterCmd)
}

func fitbitRun(cmd *cobra.Command, args []string) {
	db := db.GetDb()
	var stepsData []float64
	stmt, err := db.Prepare(`SELECT steps FROM StepsRecords WHERE date(time, 'unixepoch', 'localtime') == ?`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(time.Now().Format("2006-01-02"))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var steps float64
		err = rows.Scan(&steps)
		if err != nil {
			log.Fatal(err)
		}
		stepsData = append(stepsData, steps)
	}

	var heartData []float64
	stmt, err = db.Prepare(`SELECT heartRate FROM HeartRateRecords WHERE date(time, 'unixepoch', 'localtime') == ?`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	rows, err = stmt.Query(time.Now().Format("2006-01-02"))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var heart float64
		err = rows.Scan(&heart)
		if err != nil {
			log.Fatal(err)
		}
		heartData = append(heartData, heart)
	}

	graph := asciigraph.Plot(
		heartData,
		asciigraph.Caption("heart rate"),
		asciigraph.Height(10),
		asciigraph.Width(100))
	fmt.Println(graph)
	fmt.Println()

	graph = asciigraph.Plot(
		stepsData,
		asciigraph.Caption("steps"),
		asciigraph.Height(10),
		asciigraph.Width(100))
	fmt.Println(graph)
	fmt.Println()

	var weightData []float64
	rows, err = db.Query("SELECT weight FROM WeightRecords WHERE date >= date('2024-01-01') ORDER BY date")
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
		asciigraph.Caption("weight"),
		asciigraph.Height(10),
	)
	fmt.Println(graph)
	// sleepData := fitbit.GetSleepToday(client)
	// sleepHours := sleepData.Summary.TotalMinutesAsleep / 60
	// sleepMinutes := sleepData.Summary.TotalMinutesAsleep % 60
	//
	// fmt.Printf("slept for %d hours %d minutes\n", sleepHours, sleepMinutes)
	//
	// activityToday := fitbit.GetActivitiesToday(client)
	// fmt.Printf("steps today: %d\n", activityToday.Summary.Steps)
	// fmt.Printf("active minutes today: %d\n", activityToday.Summary.VeryActiveMinutes)
}

func fitbitLoad(cmd *cobra.Command, args []string) {
	db := db.GetDb()
	loadFitbitDb(client, db, time.Now())
}

func fitbitApi(cmd *cobra.Command, args []string) {
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
	oz, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal(err)
	}
	fitbit.AddWater(client, oz)
}
