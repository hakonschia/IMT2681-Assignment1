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
	dbURL string = "mongodb://" + dbUser + ":" + dbPassword + "@ds125502.mlab.com:25502/paragliding" // The URL used to connect to the database
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
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}

	default:
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
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
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
			return
		}

	case 2, 3: // PATH: /<id> or /<id>/<field>
		HandlerTrackFieldID(w, r)

	default: // More than 3 parts in the url (after /api/) is not implemented
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
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
					http.Error(w, "Invalid field given", http.StatusNotFound)
				}
			}

		} else {
			http.Error(w, "Invalid ID given", http.StatusNotFound)

		}

	default:
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}

/*
HandlerTicker handles /paragliding/api/ticker/
*/
func HandlerTicker(w http.ResponseWriter, r *http.Request) {

}

/*
HandlerTickerLatest handles /paragliding/api/ticker/latest/
*/
func HandlerTickerLatest(w http.ResponseWriter, r *http.Request) {

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

			var wh Webhook
			wh.URL = contentMap["webhookURL"].(string)
			wh.MinTriggerValue = int(contentMap["minTriggerValue"].(float64))
			wh.ID = nextWBID

			if webhookDB.Add(wh) {
				fmt.Fprintln(w, "ID for the new Webhook:", wh.ID)
				nextWBID++
				http.Error(w, http.StatusText(http.StatusCreated), http.StatusCreated)
			} else {
				fmt.Fprintln(w, "Couldn't add that webhook")
			}

		default:
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		}

	case 2:
		HandlerWebhookID(w, r)

	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)

	}
}
