package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/hakonschia/igcinfo_api/igcinfo"
	igc "github.com/marni/goigc"
)

var (
	startTime time.Time            // The start time of the application/API
	tracks    map[string]igc.Track // Maps ID to its corresponding track. igc.Track.Header.UniqueID is used as the key
)

func init() {
	startTime = time.Now()
	tracks = make(map[string]igc.Track)
}

//
// ----------------------------------------
//

func main() {
	port, portOk := os.LookupEnv("PORT")
	if !portOk {
		port = "8080" // 8080 is used as the default port
	}

	fmt.Println("Port is:", port)

	http.HandleFunc("/igcinfo/api/igc/", handlerAPIIGC)
	http.HandleFunc("/igcinfo/api/", handlerAPI)
	http.HandleFunc("/igcinfo/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not allowed at /igcinfo.", http.StatusNotFound)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not allowed at root.", http.StatusNotFound)
	})

	err := http.ListenAndServe(":"+port, nil)

	log.Fatalf("Server error: %s", err)
}

//
// ----------------------------------------
//

// Removes empty strings from an array
func removeEmpty(arr []string) []string {
	var newArr []string
	for _, str := range arr {
		if str != "" {
			newArr = append(newArr, str)
		}
	}

	return newArr
}

// Handles "/igcinfo/api"
func handlerAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	info := igcinfo.APIInfo{
		Uptime:  igcinfo.FormatISO8601(time.Since(startTime)),
		Info:    "Service for IGC tracks",
		Version: "V1",
	}

	json.NewEncoder(w).Encode(&info)
}

// Handles "/igcinfo/api/igc"
func handlerAPIIGC(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	parts := strings.Split(r.URL.Path, "/")

	// Remove "[ igcinifo api]" to make it more natural to work with "[igc]" being the start of the array
	parts = removeEmpty(parts[3:]) // Remove the empty strings, this makes it so "/igc/" and "/igc" is treated as the same

	switch len(parts) {
	case 1: // PATH: /igc/
		switch r.Method {
		case "GET":
			IDs := reflect.ValueOf(tracks).MapKeys() // Get the keys of the map
			fmt.Fprintln(w, IDs)

		case "POST":
			bodyStr, _ := ioutil.ReadAll(r.Body)

			urlMap := make(map[string]string)
			json.Unmarshal(bodyStr, &urlMap)

			url := urlMap["url"]
			if url == "" { // If the field name from the json is wrong no element will be found
				http.Error(w, "Invalid POST field given", http.StatusNotFound)
				return
			}

			newTrack, err := igc.ParseLocation(url)
			if err != nil { // If the passed URL couldn't be parsed the function aborts
				http.Error(w, "Invalid URL given", http.StatusNotFound)
				return
			}

			newID := newTrack.Header.UniqueID
			tracks[newID] = newTrack // Map the uniqueID to the track

			data := make(map[string]string)
			data["id"] = newID // Map the key "id" to the newly assigned ID

			json.NewEncoder(w).Encode(data) // Encode the map as a JSON object

		default: // Only POST and GET methods are implemented, any other type aborts
			http.Error(w, "Method not implemented", http.StatusNotImplemented)
			return
		}

	case 2, 3: // /<id> or /<id>/<field>
		handlerAPIIDField(w, r)

	default: // More than 3 parts in the url (after /api/) is not implemented
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}

// Handles /igcinfo/api/igc/<ID> and /igcinfo/api/igc/<id>/<field>
func handlerAPIIDField(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	parts = removeEmpty(parts[4:])

	id := parts[0]
	if track, ok := tracks[id]; ok { // The track exists
		track.Task.Start = track.Points[0] // Set the points of the track
		track.Task.Finish = track.Points[len(track.Points)-1]
		track.Task.Turnpoints = track.Points[1 : len(track.Points)-1]

		tInfo := igcinfo.TrackInfo{ // Copy the relevant information into a TrackInfo object
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
				http.Error(w, "Invalid field given", http.StatusBadRequest)
			}
		}
	} else { // ID/track was not found
		http.Error(w, "Invalid ID given", http.StatusBadRequest)
	}
}
