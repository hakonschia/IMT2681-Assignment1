package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
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

// Glider contains information about a glider, the pilot and track length
type Glider struct {
	HDate       string  `json:"H_date"`
	Pilot       string  `json:"pilot"`
	Glider      string  `json:"glider"`
	GliderID    int     `json:"glider_id"`
	TrackLength float32 `json:"track_length"`
}

// APIInfo contains basic information about the API
type APIInfo struct {
	Uptime  string `json:"uptime"` // TODO: Convert to string and match the ISO 8601 format
	Info    string `json:"info"`
	Version string `json:"version"`
}

// TrackInfo contains basic information about a track
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

// HandlerAPI handles "/igcinfo/api"
func HandlerAPI(w http.ResponseWriter, r *http.Request) {
	parts := RemoveEmpty(strings.Split(r.URL.Path, "/"))
	if len(parts) == 2 {
		w.Header().Set("content-type", "application/json")

		info := APIInfo{
			Uptime:  FormatISO8601(time.Since(startTime)),
			Info:    "Service for IGC tracks",
			Version: "V1",
		}

		json.NewEncoder(w).Encode(&info)
	} else { // /igcinfo/api/<rubbish>
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

// HandlerIGC handles "/igcinfo/api/igc"
func HandlerIGC(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	parts := strings.Split(r.URL.Path, "/")

	// Remove "[ igcinifo api]" to make it more natural to work with "[igc]" being the start of the array
	parts = RemoveEmpty(parts[3:]) // Remove the empty strings, this makes it so "/igc/" and "/igc" is treated as the same

	switch len(parts) {
	case 1: // PATH: /igc/
		switch r.Method {
		case "GET":
			var IDs []int
			for key := range tracks {
				IDs = append(IDs, key)
			}
			fmt.Fprintln(w, IDs)

		case "POST":
			bodyStr, _ := ioutil.ReadAll(r.Body) // Read the entire body (SHOULD be of form {"url": <url>})

			urlMap := make(map[string]string) // Convert the JSON string to a map
			json.Unmarshal(bodyStr, &urlMap)

			url := urlMap["url"]
			if url == "" { // If the field name from the json is wrong no element (empty string) will be returned
				http.Error(w, "Invalid POST field given", http.StatusNotFound)
				return
			}

			newTrack, err := igc.ParseLocation(url)
			if err != nil { // If the passed URL couldn't be parsed the function aborts
				http.Error(w, fmt.Sprintf("Invalid URL given: %s", err), http.StatusNotFound)
				return
			}

			if id, added := TrackAlreadyAdded(newTrack); added { // TODO: Find status code for duplicate entries
				w.Header().Set("content-type", "text/plain")
				fmt.Fprintf(w, "That track has already been added (id: %d)\n", id)
			} else {
				//newID := newTrack.Header.UniqueID
				//tracks[newID] = newTrack // Map the uniqueID to the track
				tracks[nextID] = newTrack

				data := make(map[string]int)
				data["id"] = nextID // Map the key "id" to the newly assigned ID
				nextID++

				json.NewEncoder(w).Encode(data) // Encode the map as a JSON object
			}

		default: // Only POST and GET methods are implemented, any other type aborts
			http.Error(w, "Method not implemented", http.StatusNotImplemented)
			return
		}

	case 2, 3: // PATH: /<id> or /<id>/<field>
		HandlerIDField(w, r)

	default: // More than 3 parts in the url (after /api/) is not implemented
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}

// HandlerIDField handles /igcinfo/api/igc/<ID> and /igcinfo/api/igc/<id>/<field>
func HandlerIDField(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	parts = RemoveEmpty(parts[4:])

	id, err := strconv.Atoi(parts[0])
	if err != nil { // Not an integer given
		http.Error(w, "Invalid ID type given", http.StatusBadRequest)
		return
	}

	if track, ok := tracks[id]; ok { // The track exists
		track.Task.Start = track.Points[0] // Set the points of the track
		track.Task.Finish = track.Points[len(track.Points)-1]
		track.Task.Turnpoints = track.Points[1 : len(track.Points)-1]

		tInfo := TrackInfo{ // Copy the relevant information into a TrackInfo object
			HDate:       track.Header.Date,
			Pilot:       track.Header.Pilot,
			GliderID:    track.Header.GliderID,
			Glider:      track.Header.GliderType,
			TrackLength: track.Task.Distance(),
		}

		if len(parts) == 1 { // /<id>, send back all information about the ID
			w.Header().Set("content-type", "application/json")
			json.NewEncoder(w).Encode(&tInfo)
		} else { // /<id>/<field>, send back only information about the given field
			w.Header().Set("content-type", "text/plain")
			jsonString, _ := json.Marshal(tInfo) // Convert the TrackInfo to a JSON string ([]byte)

			var trackFields map[string]interface{}   // Create a map out of the JSON string (the field is the key). Map to interface to allow all types
			json.Unmarshal(jsonString, &trackFields) // Unmarshaling the JSON string to a map

			field := parts[1]
			if res := trackFields[field]; res != nil { // If no matches were found (unknown field entered), res will be set to nil
				fmt.Fprintln(w, res)
			} else {
				http.Error(w, "Invalid field given", http.StatusNotFound)
			}
		}
	} else { // ID/track was not found
		http.Error(w, "Invalid ID given", http.StatusNotFound)
	}
}
