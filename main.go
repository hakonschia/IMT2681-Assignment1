package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	igc "github.com/marni/goigc"
)

var (
	startTime      time.Time // The start time of the application
	numericPath, _ = regexp.Compile("[0-9]")
	tracks         []igc.Track // The tracks retrieved by the user
	trackIDs       []int       // The IDs of the tracks
	lastID         int         // Last used ID
)

type jsonURL struct {
	URL string `json:"url"`
}

func init() {
	startTime = time.Now()
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	fmt.Println("Port is:", port)

	//http.HandleFunc("/igcinfo/api/igc")
	http.HandleFunc("/igcinfo/api/igc", handlerAPIIGC)
	http.HandleFunc("/igcinfo/api", handlerAPI)
	http.HandleFunc("/igcinfo/", handlerIGCINFO)
	http.HandleFunc("/", handlerRoot)

	err := http.ListenAndServe(":"+port, nil)

	log.Fatalf("Server error: %s", err)
}

// Handles root errors (no path)
func handlerRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Println("NICE BRUR")
	http.Error(w, "Not allowed at root.", http.StatusNotFound)
}

// Handles only when IGC is the path (basically only error handling)
func handlerIGCINFO(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not allowed at /igcinfo.", http.StatusNotFound)
	/*
		if len(parts) == 2 { // PATH: "/igcinfo/" (the browser will add a backslash at the end)
			http.Error(w, "Not allowed at /igcinfo.", http.StatusNotFound)
			return
		}

		//                          //
		/* ---------------------------
		 	PATH: /igcinfo/api/...
		--------------------------- */
	//                          //
	/*
		// Remove the first part of the url (/igcinfo) to make it more natural
		// to work from the standpoint of "/api/" being the root
		parts = parts[2:]

		fmt.Fprintln(w, parts)

		switch len(parts) {
		case 1: // PATH: "/api"
			handlerAPI(w, r)
		case 2: // PATH "/api/.."
			if parts[1] == "" { // "/api/"
				handlerAPI(w, r)
			} else { // "/api/.."
				if numericPath.MatchString(parts[1]) {
					handlerAPIID(w, r)
				} else {
					handlerAPIIGC(w, r)
				}
			}
		default:
			http.Error(w, "WTF R U DING KIDDO", http.StatusNotFound)
			return
		}

	*/
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
	parts = parts[3:]

	switch len(parts) {
	case 1: // PATH: /igc
		switch r.Method {
		case "GET":
			json.NewEncoder(w).Encode(&trackIDs)
		case "POST":
			var url jsonURL
			json.NewDecoder(r.Body).Decode(&url)

			//var url2 string
			//json.Unmarshal()

			newTrack, _ := igc.Parse(url.URL)
			tracks = append(tracks, newTrack)
			trackIDs = append(trackIDs, lastID+1)
			lastID++
		default: // Only POST and GET methods are implemented, any other type aborts
			return
		}
	case 2: // PATH: /igc/..
		if numericPath.MatchString(parts[2]) { // PATH: /igc/<ID>
			// Return id
		}

	default:
		http.Error(w, "WTF R U DING KIDDO=(", http.StatusNotFound)
		return
	}
}

func handlerAPIID(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
}
