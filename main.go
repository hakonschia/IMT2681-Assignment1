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
	fmt.Println("Testing branch:)")
	http.HandleFunc("/igcinfo/api/igc/", igcapi.HandlerIGC)
	http.HandleFunc("/igcinfo/api/", igcapi.HandlerAPI)
	http.HandleFunc("/igcinfo/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not allowed at /igcinfo.", http.StatusNotFound)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not allowed at root.", http.StatusNotFound)
	})

	err := http.ListenAndServe(":"+port, nil)

	log.Fatalf("Server error: %s", err)
}

//
// ----------------------------------------
//
