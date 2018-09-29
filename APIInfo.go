package main

// APIInfo contains basic information about the API
type APIInfo struct {
	Uptime  float64 `json:"uptime"` // TODO: Convert to string and match the ISO 8601 format
	Info    string  `json:"info"`
	Version string  `json:"version"`
}
