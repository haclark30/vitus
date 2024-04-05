package fitbit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/stackus/dotenv"
	"golang.org/x/oauth2"
)

const fitbitUrl = "https://api.fitbit.com"

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Expiry       string `json:"expiry"`
}

const expiryFmt = "2006-01-02T15:04:05Z07:00"

type Activity struct {
	ActivityId           int       `json:"activityId"`
	ActivityParentId     int       `json:"activityParentId"`
	ActivityParentName   string    `json:"activityParentName"`
	Calories             int       `json:"calories"`
	Description          string    `json:"description"`
	Duration             int       `json:"duration"`
	HasActiveZoneMinutes bool      `json:"hasActiveZoneMinutes"`
	HasStartTime         bool      `json:"hasStartTime"`
	IsFavorite           bool      `json:"isFavorite"`
	LastModified         time.Time `json:"lastModified"`
	LogId                int       `json:"logId"`
	Name                 string    `json:"name"`
	StartDate            string    `json:"startDate"`
	StartTime            string    `json:"startTime"`
	Steps                int       `json:"steps"`
}

type Goals struct {
	ActiveMinutes int     `json:"activeMinutes"`
	CaloriesOut   int     `json:"caloriesOut"`
	Distance      float64 `json:"distance"`
	Steps         int     `json:"steps"`
}

type Distance struct {
	Activity string  `json:"activity"`
	Distance float64 `json:"distance"`
}

type HeartRateZone struct {
	CaloriesOut float64 `json:"caloriesOut"`
	Max         int     `json:"max"`
	Min         int     `json:"min"`
	Minutes     int     `json:"minutes"`
	Name        string  `json:"name"`
}

type Summary struct {
	ActiveScore          int             `json:"activeScore"`
	ActivityCalories     int             `json:"activityCalories"`
	CaloriesBMR          int             `json:"caloriesBMR"`
	CaloriesOut          int             `json:"caloriesOut"`
	Distances            []Distance      `json:"distances"`
	FairlyActiveMinutes  int             `json:"fairlyActiveMinutes"`
	HeartRateZones       []HeartRateZone `json:"heartRateZones"`
	LightlyActiveMinutes int             `json:"lightlyActiveMinutes"`
	MarginalCalories     int             `json:"marginalCalories"`
	RestingHeartRate     int             `json:"restingHeartRate"`
	SedentaryMinutes     int             `json:"sedentaryMinutes"`
	Steps                int             `json:"steps"`
	VeryActiveMinutes    int             `json:"veryActiveMinutes"`
}

type FitnessData struct {
	Activities []Activity `json:"activities"`
	Goals      Goals      `json:"goals"`
	Summary    Summary    `json:"summary"`
}

type ActivitiesHeart struct {
	DateTime string `json:"dateTime"`
	Value    struct {
		CustomHeartRateZones []interface{}   `json:"customHeartRateZones"`
		HeartRateZones       []HeartRateZone `json:"heartRateZones"`
		RestingHeartRate     int             `json:"restingHeartRate"`
	} `json:"value"`
}

type IntradayData struct {
	Time  string `json:"time"`
	Value int    `json:"value"`
}

type ActivitiesHeartIntraday struct {
	Dataset         []IntradayData `json:"dataset"`
	DatasetInterval int            `json:"datasetInterval"`
	DatasetType     string         `json:"datasetType"`
}

type HeartRateData struct {
	ActivitiesHeart         []ActivitiesHeart       `json:"activities-heart"`
	ActivitiesHeartIntraday ActivitiesHeartIntraday `json:"activities-heart-intraday"`
}
type ActivitySteps struct {
	DateTime string `json:"dateTime"`
	Value    string `json:"value"`
}

type IntradayDataset struct {
	Dataset         []IntradayData `json:"dataset"`
	DatasetInterval int            `json:"datasetInterval"`
	DatasetType     string         `json:"datasetType"`
}
type StepsData struct {
	ActivitiesSteps      []ActivitySteps `json:"activities-steps"`
	ActivitiesStepsIntra IntradayDataset `json:"activities-steps-intraday"`
}
type WeightLog struct {
	BMI    float64 `json:"bmi"`
	Date   string  `json:"date"`
	Fat    float64 `json:"fat"`
	LogID  int64   `json:"logId"`
	Source string  `json:"source"`
	Time   string  `json:"time"`
	Weight float64 `json:"weight"`
}

type WeightData struct {
	Weight []WeightLog `json:"weight"`
}

func LoadToken() (*oauth2.Token, error) {
	file, err := os.Open("token.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var token Token
	if err := json.NewDecoder(file).Decode(&token); err != nil {
		return nil, err
	}

	tokenExpiry, err := time.Parse(expiryFmt, token.Expiry)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       tokenExpiry,
	}, nil
}
func saveToken(token *oauth2.Token) {
	file, err := os.Create("token.json")
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	t := Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry.Format(expiryFmt),
	}

	json.NewEncoder(file).Encode(t)

	if err != nil {
		log.Fatal(err)
	}
}
func NewFitbitClient() *http.Client {

	dotenv.Load()
	conf := &oauth2.Config{
		ClientID:     os.Getenv("FITBIT_CLIENT_ID"),
		ClientSecret: os.Getenv("FITBIT_API_KEY"),
		Scopes:       []string{"activity", "profile", "sleep", "weight", "heartrate", "settings", "nutrition"},
		RedirectURL:  "http://localhost:8080/fitbitCallback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.fitbit.com/oauth2/authorize",
			TokenURL: "https://api.fitbit.com/oauth2/token",
		},
	}
	ctx := context.Background()
	token, err := LoadToken()

	if err != nil || token == nil {
		verifier := oauth2.GenerateVerifier()

		url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
		fmt.Printf("visit url for auth: %v\n", url)
		fmt.Print("Enter auth code: ")
		var code string
		if _, err := fmt.Scan(&code); err != nil {
			log.Fatal(err)
		}
		token, err = conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		saveToken(token)
	}

	if !token.Valid() {
		token, err = conf.TokenSource(ctx, token).Token()
		if err != nil {
			log.Fatal(err)
		}
		saveToken(token)
	}
	client := conf.Client(ctx, token)
	return client
}

func GetActivitiesToday(fitbitClient *http.Client) *FitnessData {

	resp, err := fitbitClient.Get(
		fmt.Sprintf("%s/1/user/-/activities/date/today.json",
			fitbitUrl))
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	act := FitnessData{}
	err = json.NewDecoder(resp.Body).Decode(&act)
	if err != nil {
		log.Fatal(err)
	}

	return &act
}

func GetHeartDay(fitbitClient *http.Client) *HeartRateData {

	resp, err := fitbitClient.Get(
		fmt.Sprintf("%s/1/user/-/activities/heart/date/today/1d/1min.json",
			fitbitUrl))
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	heartrateData := HeartRateData{}
	err = json.NewDecoder(resp.Body).Decode(&heartrateData)
	if err != nil {
		log.Fatal(err)
	}

	return &heartrateData
}

func GetStepsDay(fitbitClient *http.Client) *StepsData {
	resp, err := fitbitClient.Get(
		fmt.Sprintf("%s/1/user/-/activities/steps/date/today/1d/1min.json",
			fitbitUrl))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	stepsData := StepsData{}
	err = json.NewDecoder(resp.Body).Decode(&stepsData)
	if err != nil {
		log.Fatal(err)
	}
	return &stepsData
}

func GetWeightWeek(fitbitClient *http.Client) *WeightData {
	date := time.Now().Format("2006-01-02")
	url := fmt.Sprintf("%s/1/user/-/body/log/weight/date/%s/30d.json", fitbitUrl, date)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("accept-language", "en_US")

	resp, err := fitbitClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	weightData := WeightData{}
	err = json.NewDecoder(resp.Body).Decode(&weightData)
	if err != nil {
		log.Fatal(err)
	}
	return &weightData
}

func AddWater(fitbitClient *http.Client, ounces int) {
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/1/user/-/foods/log/water.json", fitbitUrl),
		nil)
	if err != nil {
		log.Fatal(err)
	}
	q := req.URL.Query()
	q.Add("date", "today")
	q.Add("amount", strconv.Itoa(ounces))
	q.Add("unit", "fl oz")
	req.URL.RawQuery = q.Encode()

	resp, err := fitbitClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}
