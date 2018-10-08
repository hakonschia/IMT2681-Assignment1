package igcinfo

import (
	"fmt"
	"time"
)

// TrackInfo contains some information about a track
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
