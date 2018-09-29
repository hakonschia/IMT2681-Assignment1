package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	igc "github.com/marni/goigc"
)

var (
	startTime      time.Time // The start time of the application/API
	numericPath, _ = regexp.Compile("[0-9]")
	tracks         []igc.Track // The tracks retrieved by the user
	//trackIDs       []int       // The IDs of the tracks	// TODO: Make ID's strings? Look at Track.Header.UniqueID (it is a string)
	//lastID         int         // Last used ID

	trackIDs map[string]igc.Track
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

// Handles only when IGC is the path (basically only error handling)
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
	//fmt.Println("Len:", len(parts))
	parts = removeEmpty(parts[3:])

	switch len(parts) {
	case 1: // PATH: /igc/
		switch r.Method {
		case "GET":
			//fmt.Println("GETTING", trackIDs)
			var IDs []string // TODO: Make this return empty array and not "null"

			fmt.Println(IDs)

			for index := range trackIDs { // Get the indexes of the map and return the new array, TODO: Find better way to do this if possible
				IDs = append(IDs, index)
			}

			json.NewEncoder(w).Encode(&IDs)
		case "POST":
			var url jsonURL
			json.NewDecoder(r.Body).Decode(&url)

			newTrack, err := igc.ParseLocation(url.URL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			tracks = append(tracks, newTrack)

			trackIDs[newTrack.Header.UniqueID] = newTrack // Map the uniqueID to the track
			//trackIDs = append(trackIDs, lastID)
			//lastID++
		default: // Only POST and GET methods are implemented, any other type aborts
			return
		}
	case 2: // PATH: /igc/../
		if numericPath.MatchString(parts[1]) { // PATH: /igc/<ID>
			ID, _ := strconv.Atoi(parts[1]) // No need for error checking, as the if statement checks for numeric values (TODO: possibly change this? Remove if and check for error)

			var tInfo TrackInfo

			tInfo.HDate = tracks[ID].Header.Date
			tInfo.Pilot = tracks[ID].Header.Pilot
			tInfo.GliderID = tracks[ID].Header.GliderID
			tInfo.Glider = tracks[ID].Header.GliderType
			tInfo.TrackLength = tracks[ID].Task.Distance()

			//fmt.Println("ID:", ID, tracks[ID])
			json.NewEncoder(w).Encode(&tInfo)

		} else {
			http.Error(w, "Invalid ID", http.StatusNotFound) // TODO: Change error code to something more fitting (perhaps)
			return
		}
	case 3:
		w.Header().Add("content-type", "text/plain")
		ID, _ := strconv.Atoi(parts[1])
		ID = ID // DONT QUESTION IT
	default:
		http.Error(w, "WTF R U DING KIDDO=()", http.StatusNotFound)
		return
	}
}

func handlerAPIID(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
}
