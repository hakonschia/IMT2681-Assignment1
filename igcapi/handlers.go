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

const (
	dbURL = "mongodb://" + "admin" + ":" + "MSQwgXZ9HXU43NB" + "@ds125502.mlab.com:25502/paragliding" // The URL used to connect to the database
)

var (
	db        TrackDB
	webhookDB WebhookDB
)

func init() {
	db = TrackDB{
		DatabaseURL:         dbURL,
		DatabaseName:        "paragliding",
		TrackCollectionName: "tracks",
	}
	db.Init()

	webhookDB = WebhookDB{
		DatabaseURL:         dbURL,
		DatabaseName:        "paragliding",
		TrackCollectionName: "webhooks",
	}
	webhookDB.Init()
}

/*
HandlerAPI handles "/paragliding/api"
*/
func HandlerAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		parts := RemoveEmpty(strings.Split(r.URL.Path, "/"))
		if len(parts) == 2 {
			w.Header().Set("content-type", "application/json")

			info := APIInfo{
				Uptime:  FormatISO8601(time.Since(startTime)),
				Info:    "Service for IGC tracks",
				Version: "V1",
			}

			json.NewEncoder(w).Encode(&info)
		} else { // /paragliding/api/<rubbish>
			statusCode := http.StatusNotFound
			http.Error(w, http.StatusText(statusCode), statusCode)
		}

	default:
		statusCode := http.StatusNotImplemented
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}

/*
HandlerTrack handles "/paragliding/api/track"
*/
func HandlerTrack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	parts := strings.Split(r.URL.Path, "/")

	// Remove "[ paragliding api]" to make it more natural to work with "[igc]" being the start of the array
	parts = RemoveEmpty(parts[3:]) // Remove the empty strings, this makes it so "/track/" and "/track" is treated as the same

	switch len(parts) {
	case 1: // PATH: /track/
		switch r.Method {
		case http.MethodGet: // Return all the IDs in use
			IDs, _ := db.GetAllIDs()
			json.NewEncoder(w).Encode(IDs)

		case http.MethodPost: // Add a new track, return its ID
			bodyStr, err := ioutil.ReadAll(r.Body) // Read the entire body (SHOULD be of form {"url": <url>})
			if err != nil {
				fmt.Println("Couldn't read the response body")
				return
			}

			urlMap := make(map[string]string) // Convert the JSON string to a map
			json.Unmarshal(bodyStr, &urlMap)

			url := urlMap["url"]
			if url == "" { // If the field name from the json is wrong no element (empty string) will be returned
				http.Error(w, "Invalid POST field given", http.StatusNotFound)
				return
			}

			parsedTrack, err := igc.ParseLocation(url)
			if err != nil { // If the passed URL couldn't be parsed the function aborts
				http.Error(w, fmt.Sprintf("Bad Request; Invalid URL given: %s", err.Error()), http.StatusBadRequest)
				return
			}

			track := TrackInfo{
				HDate:          parsedTrack.Date,
				Pilot:          parsedTrack.Pilot,
				Glider:         parsedTrack.GliderType,
				GliderID:       parsedTrack.GliderID,
				TrackLength:    parsedTrack.Task.Distance(),
				TrackSourceURL: url,
				ID:             nextID,
				Timestamp:      time.Now().Unix(),
			}

			if db.Add(track) {
				idMap := make(map[string]int)
				idMap["id"] = nextID
				nextID++
				json.NewEncoder(w).Encode(idMap) // Encode the map as a JSON object
			} else {
				w.Header().Set("content-type", "text/plain")
				//http.Error(w, fmt.Sprintf("2%s", "f"), 201)
				fmt.Fprintf(w, "Couldn't add the track: %s\n", "xd")
			}

		default:
			statusCode := http.StatusNotImplemented
			http.Error(w, http.StatusText(statusCode), statusCode)
		}

	case 2, 3: // PATH: /<id> or /<id>/<field>
		HandlerTrackFieldID(w, r)

	default: // More than 3 parts in the url (after /api/) is not implemented
		statusCode := http.StatusBadRequest
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}

/*
HandlerTrackFieldID handles /paragliding/api/track/<ID> and /paragliding/api/track/<id>/<field>
*/
func HandlerTrackFieldID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	parts = RemoveEmpty(parts[4:])

	switch r.Method {
	case http.MethodGet:
		id, err := strconv.Atoi(parts[0])
		if err != nil { // Not an integer given
			http.Error(w, "Invalid ID type given", http.StatusBadRequest)
			return
		}

		track, found := db.Get(id)
		if found {
			response := make(map[string]interface{})
			response["H_date"] = track.HDate
			response["pilot"] = track.Pilot
			response["glider"] = track.Glider
			response["glider_id"] = track.Glider
			response["track_length"] = track.TrackLength
			response["track_src_url"] = track.TrackSourceURL

			fmt.Println(track.HDate)
			if len(parts) == 1 { // /track/<ID>/
				json.NewEncoder(w).Encode(track)
			} else { // /track/<ID>/<field>/
				w.Header().Set("content-type", "text/plain")

				field := parts[1]
				if res, found := response[field]; found {
					fmt.Fprintln(w, res)
				} else {
					http.Error(w, "Invalid field given", http.StatusBadRequest)
				}
			}

		} else {
			http.Error(w, "Invalid ID given", http.StatusBadRequest)

		}

	default:
		statusCode := http.StatusNotImplemented
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}

/*
HandlerTicker handles /paragliding/api/ticker/
*/
func HandlerTicker(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("content-type", "application/json")
		parts := RemoveEmpty(strings.Split(r.URL.Path, "/"))
		parts = parts[2:]

		pagingSize := 5

		taskStart := time.Now().Unix()

		tracks, err := db.GetAll()
		if err != nil {
			fmt.Println("Error retrieving from DB, DB could be empty.")
			return
		}

		type ticker struct {
			TLatest    int64 `json:"t_latest"`
			TStart     int64 `json:"t_start"`
			TStop      int64 `json:"t_stop"`
			Tracks     []int `json:"tracks"`
			Processing int64 `json:"processing"`
		}

		switch len(parts) {
		case 1, 2: // /ticker/ and /ticker/<timestamp>
			if len(parts) == 2 { // If /ticker/<timestamp> we simply modify the start of the track array
				timestamp, err := strconv.Atoi(parts[1])
				if err != nil {
					http.Error(w, "Invalid ID type given", http.StatusBadRequest)
					return
				}

				trackStart := -1
				for i, val := range tracks {
					if val.Timestamp == int64(timestamp) {
						trackStart = i + 1 // The tracks should start at the NEXT after the timestamp
					}
				}

				if trackStart == -1 { // If the timestamp wasn't found
					http.Error(w, "Invalid ID given", http.StatusBadRequest)
					return
				}

				if trackStart >= len(tracks) { // The timestamp given is the newest in the DB
					w.Header().Set("content-type", "text/plain")
					fmt.Fprintln(w, "No new added tracks")
					return
				}

				tracks = tracks[trackStart:] // Remove the tracks before
			}

			trackIDs := []int{}
			amountOfTracks := Min(pagingSize, len(tracks)) // The amount of tracks on the "page"

			for i := 0; i < amountOfTracks; i++ {
				trackIDs = append(trackIDs, tracks[i].ID)
			}

			response := ticker{
				TLatest:    tracks[len(tracks)-1].Timestamp,
				TStart:     tracks[0].Timestamp,
				TStop:      tracks[amountOfTracks-1].Timestamp,
				Tracks:     trackIDs,
				Processing: time.Now().Unix() - taskStart,
			}

			json.NewEncoder(w).Encode(response)

		default:
			statusCode := http.StatusBadRequest
			http.Error(w, http.StatusText(statusCode), statusCode)
		}

	default:
		statusCode := http.StatusNotImplemented
		http.Error(w, http.StatusText(statusCode), statusCode)
	}

}

/*
HandlerTickerLatest handles /paragliding/api/ticker/latest/
*/
func HandlerTickerLatest(w http.ResponseWriter, r *http.Request) {
	parts := RemoveEmpty(strings.Split(r.URL.Path, "/"))
	if len(parts) == 4 {
		t, err := db.GetLast()
		if err != nil {
			fmt.Println("Couldn't get a track :)")
			return
		}

		w.Header().Set("content-type", "text/plain")
		fmt.Fprintln(w, t.Timestamp)
	} else {
		statusCode := http.StatusNotFound
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}

/*
HandlerWebhook handles /paragliding/api/webhook/..
*/
func HandlerWebhook(w http.ResponseWriter, r *http.Request) {
	parts := RemoveEmpty(strings.Split(r.URL.Path, "/"))
	parts = parts[3:]

	switch len(parts) {
	case 1:
		switch r.Method {
		case http.MethodPost:
			bodyStr, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Println("Couldn't read the response body")
				return
			}

			contentMap := make(map[string]interface{})
			json.Unmarshal(bodyStr, &contentMap)

			wh := Webhook{
				URL:       contentMap["webhookURL"].(string),
				ID:        nextWBID,
				Timestamp: time.Now().Unix(),
			}

			triggerValue := contentMap["minTriggerValue"]

			if triggerValue == nil || triggerValue.(int) == 0 {
				triggerValue = 1
			}

			wh.MinTriggerValue = triggerValue.(int)

			if webhookDB.Add(wh) {
				fmt.Fprintln(w, "ID for the new Webhook:", wh.ID)
				nextWBID++
				http.Error(w, http.StatusText(http.StatusCreated), http.StatusCreated)
			} else {
				fmt.Fprintln(w, "Couldn't add that webhook")
			}

		default:
			statusCode := http.StatusNotImplemented
			http.Error(w, http.StatusText(statusCode), statusCode)
		}

	case 2:
		HandlerWebhookID(w, r)

	default:
		statusCode := http.StatusBadRequest
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}

/*
HandlerWebhookID handles /webhook/new_track/<webhook_id>
*/
func HandlerWebhookID(w http.ResponseWriter, r *http.Request) {
	parts := RemoveEmpty(strings.Split(r.URL.Path, "/"))
	parts = parts[4:]

	ID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid ID type given", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		wh := webhookDB.Get(ID)
		json.NewEncoder(w).Encode(wh)

	case http.MethodDelete:
		wh := webhookDB.Delete(ID)
		json.NewEncoder(w).Encode(wh)

	default:
		statusCode := http.StatusNotImplemented
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}

/*
HandlerAdminTrackCount handles /paragliding/admin/api/tracks_count/
*/
func HandlerAdminTrackCount(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("content-type", "text/plain")
		fmt.Fprintln(w, db.Count())

	default:
		statusCode := http.StatusNotImplemented
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}

/*
HandlerAdminTrack handles /paragliding/admin/api/tracks/
*/
func HandlerAdminTrack(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		w.Header().Set("content-type", "text/plain")
		countDeleted := db.DeleteAll()
		fmt.Fprintln(w, "Deleted tracks:", countDeleted)

	default:
		statusCode := http.StatusNotImplemented
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}
