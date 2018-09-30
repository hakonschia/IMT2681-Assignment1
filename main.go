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
	trackIDs  map[string]igc.Track // The uniqueID in igc.track is used as the indexing
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

	case 3: // PATH: /igc/<id>/<field> // TODO: Fix shitty shit code
		if track, ok := trackIDs[parts[1]]; ok { // If the ID was found in "trackIDs", it exists
			w.Header().Add("content-type", "text/plain")

			//marshalled, _ := json.Marshal(track)
			//fmt.Println(string(marshalled))
			var tInfo TrackInfo

			tInfo.HDate = track.Header.Date
			tInfo.Pilot = track.Header.Pilot
			tInfo.GliderID = track.Header.GliderID
			tInfo.Glider = track.Header.GliderType
			tInfo.TrackLength = track.Task.Distance()

			jsonString, _ := json.Marshal(tInfo) // Convert the TrackInfo to a json string
			//fmt.Println(string(jsonString))

			var data map[string]interface{} // Create a map out of the json string (the json field is the index). Map to interface to allow all types
			json.Unmarshal([]byte(jsonString), &data)

			res := data[parts[2]]
			fmt.Println(parts[2], ":", res)
			/*
				switch parts[2] {
				case "pilot":
					fmt.Println(track.Header.Pilot)
					//fmt.Fprintln(w, track.Header.Pilot)
					json.NewEncoder(w).Encode(track.Header.Pilot) // Problems with fmt.Fprint, get "Expected '<first_letter>'" error in postman/browser
				case "glider":
					fmt.Fprintln(w, track.Header.GliderType)
				case "glider_id":
					fmt.Fprintln(w, track.Header.GliderID)
				case "track_length":
					fmt.Fprintln(w, track.Task.Distance())
				case "H_date":
					fmt.Fprintln(w, track.Header.Date)
				default:
					http.Error(w, "Unknown field given", http.StatusNotFound)
					return
				}

			*/
		} else {
			http.Error(w, "Invalid ID", http.StatusNotFound) // TODO: Change error code to something more fitting (perhaps)
			return
		}

	case 2: // PATH: /igc/<id>/
		if track, ok := trackIDs[parts[1]]; ok {
			var tInfo TrackInfo

			tInfo.HDate = track.Header.Date
			tInfo.Pilot = track.Header.Pilot
			tInfo.GliderID = track.Header.GliderID
			tInfo.Glider = track.Header.GliderType
			tInfo.TrackLength = track.Task.Distance()

			jsonString, _ := json.Marshal(tInfo)
			fmt.Println(string(jsonString))

			var dat map[string]string
			json.Unmarshal([]byte(jsonString), &dat)

			invoices := dat["pilot"]
			fmt.Println(invoices)

			json.NewEncoder(w).Encode(&tInfo)
		} else {
			http.Error(w, "Invalid ID", http.StatusNotFound) // TODO: Change error code to something more fitting (perhaps)
			return
		}

	default:
		http.Error(w, "WTF R U DING KIDDO=()", http.StatusNotFound)
		return
	}
}

func handlerAPIID(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
}
