package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hakonschia/igcinfo_api/igcapi"
)

//
// ----------------------------------------
//

func main() {
	port, portOk := os.LookupEnv("PORT")
	if !portOk {
		port = "8080" // 8080 is used as the default port
	}

	fmt.Println("Port is:", port)

	http.HandleFunc("/paragliding/api/ticker/latest/", igcapi.HandlerTrack)
	http.HandleFunc("/paragliding/api/ticker/", igcapi.HandlerTicker)
	http.HandleFunc("/paragliding/api/track/", igcapi.HandlerTrack)
	http.HandleFunc("/paragliding/api/", igcapi.HandlerAPI)
	http.HandleFunc("/paragliding/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/paragliding/api/", http.StatusMovedPermanently)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	err := http.ListenAndServe(":"+port, nil)

	log.Fatalf("Server error: %s", err)
}

//
// ----------------------------------------
//
