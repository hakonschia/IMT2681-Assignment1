package igcapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	igc "github.com/marni/goigc"
)

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
		case "GET": // Return all the IDs in use
			IDs := []int{}
			for key := range tracks {
				IDs = append(IDs, key)
			}
			json.NewEncoder(w).Encode(IDs)

		case "POST": // Add a new track, return its ID
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
				tracks[nextID] = newTrack

				data := make(map[string]int) // A map for the JSON response
				data["id"] = nextID
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

	// By using a map instead of slice/array for the tracks, it's very easy to check if it exists
	// since you can get a boolean value back as well as the object from a map (or check if val != nil)
	// Although not implemented here, this is very handy for deleting tracks, without having to
	// keeping track of all the IDs which are deleted

	if track, ok := tracks[id]; ok { // The track exists
		track.Task.Start = track.Points[0] // Set the points of the track
		track.Task.Finish = track.Points[len(track.Points)-1]
		track.Task.Turnpoints = track.Points[1 : len(track.Points)-1] // [from, including : to, not including]

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
			if res, found := trackFields[field]; found {
				fmt.Fprintln(w, res)
			} else {
				http.Error(w, "Invalid field given", http.StatusNotFound)
			}
		}
	} else { // ID/track was not found
		http.Error(w, "Invalid ID given", http.StatusNotFound)
	}
}
