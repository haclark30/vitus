package fitbit

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type SleepSummary struct {
	TotalMinutesAsleep int `json:"totalMinutesAsleep"`
}

type SleepData struct {
	Summary SleepSummary `json:"summary"`
}

func GetSleepToday(fitbitClient *http.Client) *SleepData {
	url := fmt.Sprintf("%s/1.2/user/-/sleep/date/today.json", fitbitUrl)
	resp, err := fitbitClient.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	sleepData := SleepData{}

	err = json.NewDecoder(resp.Body).Decode(&sleepData)
	if err != nil {
		log.Fatal(err)
	}
	return &sleepData
}
