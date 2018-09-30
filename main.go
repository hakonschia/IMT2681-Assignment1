package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	igc "github.com/marni/goigc"
)

var (
	startTime time.Time            // The start time of the application/API
	tracks    []igc.Track          // The tracks retrieved by the user
	trackIDs  map[string]igc.Track // The uniqueID in igc.Track.Header is used as the key
)

type jsonURL struct {
	URL string `json:"url"`
}

func init() {
	startTime = time.Now()
	trackIDs = make(map[string]igc.Track)
}

func main() {
	port, portOk := os.LookupEnv("PORT")
	if !portOk {
		port = "8080" // 8080 is used as the default port
	}

	fmt.Println("Port is:", port)

	http.HandleFunc("/igcinfo/api/igc/", handlerAPIIGC)
	http.HandleFunc("/igcinfo/api/", handlerAPI)
	http.HandleFunc("/igcinfo/", handlerIGCINFO)
	http.HandleFunc("/", handlerRoot)

	err := http.ListenAndServe(":"+port, nil)

	log.Fatalf("Server error: %s", err)
}

// Removes empty strings from a given array
func removeEmpty(arr []string) []string {
	var newArr []string
	for _, str := range arr {
		if str != "" {
			newArr = append(newArr, str)
		}
	}

	return newArr
}

// Handles root errors (no path)
func handlerRoot(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not allowed at root.", http.StatusNotFound)
}

// Handles /igcinfo/ (error handling)
func handlerIGCINFO(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not allowed at /igcinfo.", http.StatusNotFound)
}

// Handles "/igcinfo/api"
func handlerAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	info := APIInfo{
		Uptime:  time.Since(startTime).String(), // Uptime TODO: Format to ISO8601
		Info:    "Service for IGC tracks",
		Version: "V1",
	}
	//ts := tds.UTC().Format("P2006-01-02T15:04:05-0700")

	json.NewEncoder(w).Encode(&info)
}

// Handles "/igcinfo/api/igc"
func handlerAPIIGC(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	parts := strings.Split(r.URL.Path, "/")

	// Remove "[ igcinifo api]" to make it more natural to work with "[igc]" being the start of the array
	parts = removeEmpty(parts[3:]) // Remove the empty strings, this makes it so "/igc/" and "/igc" is treated as the same

	switch len(parts) {
	case 1: // PATH: /igc/
		switch r.Method {
		case "GET":
			var IDs []string

			for key := range trackIDs { // Put the keys of the map in the new array
				IDs = append(IDs, key)
			}

			fmt.Fprintln(w, IDs) // This returns an empty array instead of null, but is this correct for "application/json"?
			//json.NewEncoder(w).Encode(&IDs)

		case "POST":
			var url jsonURL
			json.NewDecoder(r.Body).Decode(&url)

			newTrack, err := igc.ParseLocation(url.URL)
			if err != nil { // If the passed URL couldn't be parsed the function aborts
				http.Error(w, "Invalid URL", http.StatusNotFound)
				return
			}

			tracks = append(tracks, newTrack)

			newID := newTrack.Header.UniqueID
			trackIDs[newID] = newTrack // Map the uniqueID to the track

			data := make(map[string]string)
			data["id"] = newID // Map the key "id" to the newly assigned ID

			json.NewEncoder(w).Encode(data) // Encode the map as a JSON object

		default: // Only POST and GET methods are implemented, any other type aborts
			return
		}

	case 2, 3: // 2 or 3 parts means /<id> or /<id>/<field>
		handlerAPIID(w, r)

	default: // More than 3 parts in the url (after /api/) is not implemented
		http.Error(w, "", http.StatusNotFound)
		return
	}
}

// Handles /igcinfo/api/igc/<ID> and /igcinfo/api/igc/<id>/<field>
func handlerAPIID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	parts = removeEmpty(parts[4:])

	if track, ok := trackIDs[parts[0]]; ok { // The track exists
		tInfo := TrackInfo{ // Copy the relevant information into a TrackInfo object
			HDate:       track.Header.Date,
			Pilot:       track.Header.Pilot,
			GliderID:    track.Header.GliderID,
			Glider:      track.Header.GliderType,
			TrackLength: track.Task.Distance(),
		}

		if len(parts) == 1 { // /<id>
			w.Header().Add("content-type", "application/json")
			json.NewEncoder(w).Encode(&tInfo)
		} else { // /<id>/<field>
			w.Header().Add("content-type", "text/plain")
			jsonString, _ := json.Marshal(tInfo) // Convert the TrackInfo to a json string

			var trackFields map[string]interface{}   // Create a map out of the json string (the json field is the index). Map to interface to allow all types
			json.Unmarshal(jsonString, &trackFields) // Unmarshaling converts the json string to a map

			field := parts[1]
			if res := trackFields[field]; res != nil { // If no matches were found (unknown field entered), res will be set to nil
				fmt.Fprintln(w, res)
			} else {
				http.Error(w, "Invalid field given", http.StatusNotFound)
			}
		}
	} else { // ID was not found
		http.Error(w, "Invalid ID given", http.StatusNotFound)
	}
}
