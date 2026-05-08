package main

import (
    "log"
    "net/http"
)

func main() {

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("conduit proxy running"))
    })

    log.Println("[conduit] listening on :8080")

    log.Fatal(http.ListenAndServe(":8080", nil))
}