package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hakonschia/igcinfo_api/packagetest"
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

	http.HandleFunc("/igcinfo/api/igc/", HandlerIGC)
	http.HandleFunc("/igcinfo/api/", HandlerAPI)
	http.HandleFunc("/igcinfo/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not allowed at /igcinfo.", http.StatusNotFound)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		packagetest.PrintHello(w)
		http.Error(w, "Not allowed at root.", http.StatusNotFound)
	})

	err := http.ListenAndServe(":"+port, nil)

	log.Fatalf("Server error: %s", err)
}

//
// ----------------------------------------
//
