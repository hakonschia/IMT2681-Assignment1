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
	trackIDs  map[string]igc.Track // The uniqueID in igc.Track.Header is used as the indexing
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
	w.Header().Add("content-type", "application/json") // Set the response type

	var info APIInfo
	//ts := tds.UTC().Format("2006-01-02T15:04:05-0700")

	info.Uptime = time.Since(startTime).String() // TODO: make ISO 8601
	info.Info = "Service for IGC tracks"
	info.Version = "V1"

	json.NewEncoder(w).Encode(&info)
}

// Handles "/igcinfo/api/igc"
func handlerAPIIGC(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	parts := strings.Split(r.URL.Path, "/")

	// Remove "[ igcinifo api]" to make it more natural to work with "[igc]" being the start of the array
	parts = removeEmpty(parts[3:]) // Remove the empty strings as well, this makes "/igc/" and "/igc" the same

	switch len(parts) {
	case 1: // PATH: /igc/
		switch r.Method {
		case "GET":
			var IDs []string // TODO: Make this return empty array and not "null"

			for index := range trackIDs { // Get the indexes of the map and return the new array, TODO: Find better way to do this if possible
				IDs = append(IDs, index)
			}

			//fmt.Fprint(w, IDs) // This returns an empty array instead of null, but is this correct for "application/json"?
			// Also it doesnt work, browser says "Expected ',' instead of 'S'"
			json.NewEncoder(w).Encode(&IDs)

		case "POST":
			var url jsonURL
			json.NewDecoder(r.Body).Decode(&url)

			newTrack, err := igc.ParseLocation(url.URL)
			if err != nil { // If the passed URL couldn't be parsed the function aborts
				http.Error(w, "Invalid URL", http.StatusNotFound)
				return
			}

			tracks = append(tracks, newTrack)

			trackIDs[newTrack.Header.UniqueID] = newTrack // Map the uniqueID to the track

		default: // Only POST and GET methods are implemented, any other type aborts
			return
		}

	case 2:
		handlerAPIID(w, r)
	case 3:
		handlerAPIID(w, r)

	default: // More than 4 parts in the url
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
			track.Header.Date,
			track.Header.Pilot,
			track.Header.GliderID,
			track.Header.GliderType,
			track.Task.Distance(),
		}

		if len(parts) == 1 { // /<id>
			w.Header().Add("content-type", "application/json")
			json.NewEncoder(w).Encode(&tInfo)
		} else { // /<id>/<field>
			w.Header().Add("content-type", "text/plain")
			jsonString, _ := json.Marshal(tInfo) // Convert the TrackInfo to a json string

			var data map[string]interface{} // Create a map out of the json string (the json field is the index). Map to interface to allow all types
			json.Unmarshal([]byte(jsonString), &data)

			if res := data[parts[1]]; res != nil { // If no matches were found, res will be set to nil
				fmt.Println(parts[1], ":", res)
				//fmt.Fprintln(w, res)
			} else {
				http.Error(w, "Invalid field given", http.StatusNotFound)
			}
		}
	} else { // ID was not found
		http.Error(w, "Invalid ID given", http.StatusNotFound)
	}
}
