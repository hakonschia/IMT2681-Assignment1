package igcapi

import (
	"fmt"
	"reflect"
	"time"

	igc "github.com/marni/goigc"
)

var (
	startTime time.Time         // The start time of the application/API
	tracks    map[int]igc.Track // Maps ID to its corresponding track
	nextID    int               // The next ID to be used
)

func init() {
	startTime = time.Now()
	tracks = make(map[int]igc.Track)
	nextID = 1 // Start at 1 to avoid potential conversions from string (which returns 0 if not an int)
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
TrackInfo contains basic information about a track
*/
type TrackInfo struct {
	HDate       time.Time `json:"H_date"`
	Pilot       string    `json:"pilot"`
	Glider      string    `json:"glider"`
	GliderID    string    `json:"glider_id"`
	TrackLength float64   `json:"track_length"`
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

// TrackAlreadyAdded checks if a track has already been added. Returns the ID if it exists
func TrackAlreadyAdded(track igc.Track) (int, bool) {
	for id, val := range tracks {
		if reflect.DeepEqual(track, val) {
			return id, true
		}
	}
	return 0, false
}
