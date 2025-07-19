package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := flag.String("port", "8080", "port to run the server on")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/readiness", readinessHandler)

	server := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	fmt.Printf("Server starting on port %s\n", *port)
	log.Fatal(server.ListenAndServe())
}