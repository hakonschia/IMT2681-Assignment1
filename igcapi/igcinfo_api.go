package igcapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	igc "github.com/marni/goigc"
)

var (
	startTime time.Time // The start time of the application/API
	nextID    int       // The next ID to be used
)

const (
	discordWebhook = "https://discordapp.com/api/webhooks/503200944650059786/oMgOMvTiNxpuVonieqfvpg5jajVZDV8I3cHQxvkT92ww4jrpBmrANvbxyFVVQXGYuCnk"
)

func init() {
	startTime = time.Now()
	nextID = db.GetLastID()
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
Track contains a track (igc.Track), its ID and the source URL
*/ // TODO: Find better name
type Track struct {
	TrackID        int    `json:"trackid"`
	TrackSourceURL string `json:"track_src_url"`
	igc.Track      `json:"igctrack"`
}

// FormatISO8601 formats time.Duration to a string according to the ISO8601 standard
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

// RemoveEmpty removes empty strings from an array
func RemoveEmpty(arr []string) []string {
	var newArr []string
	for _, str := range arr {
		if str != "" {
			newArr = append(newArr, str)
		}
	}

	return newArr
}

// ClockTrigger checks every 10 minutes if the number of tracks has changed, and notifies if it has
func ClockTrigger() {
	delay := time.Minute * 10
	previousTrackNum := db.Count()

	for {
		if previousTrackNum != db.Count() {
			previousTrackNum = db.Count()
			NotifyDiscord()
		}

		time.Sleep(delay)
	}

}

// NotifyDiscord sends an update to a discord channel
func NotifyDiscord() {
	content := make(map[string]string)
	content["content"] = "There has been an update to the database!\n"
	raw, _ := json.Marshal(content)

	resp, err := http.Post(discordWebhook, "application/json", bytes.NewBuffer(raw))
	if err != nil {
		fmt.Println(err)
		fmt.Println(ioutil.ReadAll(resp.Body))
	}
}

func SendNotificationAnyway() {
	content := make(map[string]string)
	content["content"] = "Sadly, no update. But hey, you deserve to be loved as well!\n"
	raw, _ := json.Marshal(content)

	resp, err := http.Post(discordWebhook, "application/json", bytes.NewBuffer(raw))
	if err != nil {
		fmt.Println(err)
		fmt.Println(ioutil.ReadAll(resp.Body))
	}
}
