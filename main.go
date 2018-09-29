package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Get port from environment
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	log.Print("Add handler to root path")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Recived %s from %s for %s", r.Method, r.RemoteAddr, r.URL)
		fmt.Fprint(w, "Hello world")
	})

	log.Printf("Setting up server to listen at port %s", port)

	// Run the server
	// This will block the current thread
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)

	// We will only get to this statement if the server unexpectedly crashes
	log.Fatalf("Server error: %s", err)
}
