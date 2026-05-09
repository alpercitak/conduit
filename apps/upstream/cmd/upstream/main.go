package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	name := os.Getenv("UPSTREAM_NAME")
	if name == "" {
		name = "upstream"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "response from %s\n", name)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("[%s] listening on :"+port, name)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
