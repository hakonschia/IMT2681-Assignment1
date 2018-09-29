package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var startTime time.Time // The start time of the application

func init() {
	startTime = time.Now()
}

// Handles root errors (no path)
func handlerRoot(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not allowed at root.", http.StatusNotFound)
}

// Handles only when IGC is the path (basically only error handling)
func handlerIGC(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/") // Split the url into all its parts

	if len(parts) == 2 || (len(parts) == 3 && parts[2] == "") { // PATH: "/igcinfo" or "/igcinfo/" (NOTE: the browser seems to automatically add the second slash for THIS path)
		http.Error(w, "Not found at root.", http.StatusNotFound)
		return
	}

	//							//
	/* ---------------------------
	 	PATH: /igcinfo/api/...
	--------------------------- */
	//							//

	// Remove the first part of the url (/igcinfo) to make it more natural
	// to work from the standpoint of "/api/" being the root
	parts = parts[2:]

	if len(parts) == 1 || (len(parts) == 2 && parts[1] == "") { // PATH: "/api/" or "/api"
		w.Header().Add("content-type", "application/json") // Set the response type

		// Provide basic information about the api

		//var uptime string // TODO: Convert to ISO 8601 format.
		// time.Since(startTime).String() might be useful (Format: 1h2m0.3s)
		var info APIInfo

		//td := time.Now().Sub(startTime)

		//ts := tds.UTC().Format("2006-01-02T15:04:05-0700")

		info.Uptime = "s" // Just get the seconds for now
		info.Info = "Service for IGC tracks"
		info.Version = "V1"

		json.NewEncoder(w).Encode(&info)
	} /*else if ( GET /api/igc ) {
		w.Header().Add("content-type", "application/json")
	} else if ( GET /api/igc/<id> ) {
		w.Header().Add("content-type", "application/json")
	} else if ( GET /api/igc/<id>/<field>) {
		w.Header().Add("content-type", "text/plain")
	} else { // /api/<rubbish>
		http.Error(w, "Nothind found at /api/", http.StatusNotFound)
		return
	} */

	// Status 200: OK
}

// This function handles everything to do with the API, NOTE: Doesn't work as everything after /igcinfo/ is treated in the same function
func handlerAPI(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")

	parts = parts[2:]
	fmt.Fprintln(w, parts)
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}
	port = ":" + port

	http.HandleFunc("/", handlerRoot)
	http.HandleFunc("/igcinfo/", handlerIGC)
	//http.HandleFunc("/igcinfo/API", handlerAPI)

	err := http.ListenAndServe(port, nil)

	log.Fatalf("Server error: %s", err)
}
