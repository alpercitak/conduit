package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/alpercitak/conduit/apps/proxy/internal/router"
	"github.com/alpercitak/conduit/apps/proxy/internal/upstream"
)

func main() {

	pool := upstream.NewPool([]*upstream.Upstream{
		{ID: "alpha", URL: "http://localhost:3001"},
		{ID: "beta", URL: "http://localhost:3002"},
	})

	r := router.New(pool)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		u, err := r.Pick(req.Context())
		if err != nil {
			http.Error(w, "no upstream available", 503)
			return
		}

		target, _ := url.Parse(u.URL)
		proxy := httputil.NewSingleHostReverseProxy(target)

		log.Printf("[conduit] %s %s -> %s", req.Method, req.URL.Path, u.URL)

		proxy.ServeHTTP(w, req)
	})

	log.Println("[conduit] proxy running on :8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
