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
		data := make([]float64, 0)
		for _, hr := range heartRate.ActivitiesHeartIntraday.Dataset {
			data = append(data, float64(hr.Value))
		}
		graph := asciigraph.Plot(
			data,
			asciigraph.Height(10),
			asciigraph.Width(100))
		fmt.Println(graph)
		fmt.Println()

		stepsData := fitbit.GetStepsDay(client)
		data = make([]float64, 0)
		for _, steps := range stepsData.ActivitiesStepsIntra.Dataset {
			data = append(data, float64(steps.Value))
		}
		graph = asciigraph.Plot(
			data,
			asciigraph.Height(10),
			asciigraph.Width(100))
		fmt.Println(graph)
	}
}
