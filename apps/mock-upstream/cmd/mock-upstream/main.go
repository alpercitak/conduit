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

	log.Printf("[%s] listening on :3001", name)

	log.Fatal(http.ListenAndServe(":3001", nil))
}
