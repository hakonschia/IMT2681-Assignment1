package igcapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

var (
	startTime time.Time
	nextID    int
	nextWBID  int
)

const (
	discordWebhookURL string = "https://discordapp.com/api/webhooks/503200944650059786/oMgOMvTiNxpuVonieqfvpg5jajVZDV8I3cHQxvkT92ww4jrpBmrANvbxyFVVQXGYuCnk"
)

func init() {
	startTime = time.Now()
	nextID = db.GetLastID() + 1
	nextWBID = webhookDB.GetLastID() + 1
}

/*
APIInfo contains basic information about the API
*/
type APIInfo struct {
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
	Version string `json:"version"`
}

/*
TrackInfo contains meta data about a track, including its source url and database ID
*/
type TrackInfo struct {
	HDate          time.Time `json:"H_date"`
	Pilot          string    `json:"pilot"`
	Glider         string    `json:"glider"`
	GliderID       string    `json:"glider_id"`
	TrackLength    float64   `json:"track_length"`
	TrackSourceURL string    `json:"track_src_url"`
	ID             int       `json:"-"`
}

/*
Webhook contains the URL and minimum trigger value for a webhook
*/
type Webhook struct {
	URL             string `json:"webhookURL"`
	MinTriggerValue int    `json:"minTriggerValue`
	ID              int    `json:"-"`
}

/*
FormatISO8601 formats time.Duration to a string according to the ISO8601 standard
*/
func FormatISO8601(t time.Duration) string {
	seconds := int64(t.Seconds()) % 60 // These functions return the total time for each field (e.g 200 seconds)
	minutes := int64(t.Minutes()) % 60 // Using modulo we get the correct values for each field
	hours := int64(t.Hours()) % 24

	totalHours := int64(t.Hours())
	days := (totalHours / 24) % 30 // Doesnt really work since it's not 30 days in each month
	months := (totalHours / (24 * 30)) % 12
	years := totalHours / (24 * 30 * 12)

	return fmt.Sprint("P", years, "Y", months, "M", days, "DT", hours, "H", minutes, "M", seconds, "S")
}

/*
RemoveEmpty removes empty strings from an array
*/
func RemoveEmpty(arr []string) []string {
	var newArr []string
	for _, str := range arr {
		if str != "" {
			newArr = append(newArr, str)
		}
	}

	return newArr
}

/*
ClockTrigger checks every 10 minutes if the number of tracks has changed, and notifies if it has
*/
func ClockTrigger() {
	delay := time.Minute * 1
	previousTrackNum := db.Count()

	for {
		if previousTrackNum != db.Count() {
			previousTrackNum = db.Count()
			NotifyDiscord()
		} else if rand.Intn(10) == 1 { // This probably should be seeded, but who cares really
			SendNotificationAnyway()
		}

		time.Sleep(delay)
	}

}

/*
NotifyDiscord sends an update to a discord channel
*/
func NotifyDiscord() {
	content := make(map[string]string)
	content["content"] = "There has been an update to the database!\n"
	raw, _ := json.Marshal(content)

	_, err := http.Post(discordWebhookURL, "application/json", bytes.NewBuffer(raw))
	if err != nil {
		fmt.Println(err.Error())
	}
}

/*
SendNotificationAnyway sends a notification to the discord server even if no
update has happened, because everyone deserves love sometimes
*/
func SendNotificationAnyway() {
	content := make(map[string]string)
	content["content"] = "Sadly, no update. But hey, you deserve to be loved anyways!\n"
	raw, _ := json.Marshal(content)

	_, err := http.Post(discordWebhookURL, "application/json", bytes.NewBuffer(raw))
	if err != nil {
		fmt.Println(err.Error())
	}
}
