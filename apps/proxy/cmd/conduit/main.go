package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {

	target, err := url.Parse("http://localhost:3001")
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[conduit] %s %s -> %s", r.Method, r.URL.Path, target)
		proxy.ServeHTTP(w, r)
	})

	log.Println("[conduit] proxy running on :8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
