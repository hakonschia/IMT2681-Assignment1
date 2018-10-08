package igcinfo

// Glider is a container with information about a glider, the pilot and track length
type Glider struct {
	HDate       string  `json:"H_date"`
	Pilot       string  `json:"pilot"`
	Glider      string  `json:"glider"`
	GliderID    int     `json:"glider_id"`
	TrackLength float32 `json:"track_length"`
}
