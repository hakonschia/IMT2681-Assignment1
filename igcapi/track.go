package igcapi

import (
	"time"

	igc "github.com/marni/goigc"
)

/*
TrackInfo contains basic information about a track
*/
type TrackInfo struct {
	HDate          time.Time `json:"H_date"`
	Pilot          string    `json:"pilot"`
	Glider         string    `json:"glider"`
	GliderID       string    `json:"glider_id"`
	TrackLength    float64   `json:"track_length"`
	TrackSourceURL string    `json:"track_src_url"`
}

/*
Track contains a track and its ID
*/ // TODO: Find better name
type Track struct {
	TrackID   int `json:"trackid"`
	igc.Track `json:"igctrack"`
}
