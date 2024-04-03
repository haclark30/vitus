package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/guptarohit/asciigraph"
	"github.com/haclark30/vitus/fitbit"
)

const fitbitUrl = "https://api.fitbit.com/"

func main() {
	args := os.Args[1:]
	client := fitbit.NewFitbitClient()
	if len(args) == 1 {
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
	} else if len(args) > 1 {
		if args[0] == "water" {
			oz, err := strconv.Atoi(args[1])
			if err != nil {
				log.Fatal(err)
			}
			fitbit.AddWater(client, oz)
		}

	} else {
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

		weightWeek := fitbit.GetWeightWeek(client)
		weightData := make([]float64, 0)

		for _, w := range weightWeek.Weight {
			weightData = append(weightData, w.Weight)
		}

		graph = asciigraph.Plot(
			weightData,
			asciigraph.Height(10),
		)
		fmt.Println(graph)
	}

}
