package metrics

import (
	"encoding/json"
	"net/http"
)

func HTTPHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		p50, p95, p99, requests, errors := store.Summary()

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(map[string]any{
			"requests_total": requests,
			"errors_total":   errors,
			"p50":            p50,
			"p95":            p95,
			"p99":            p99,
		})

		if err != nil {
			http.Error(
				w,
				"failed to encode metrics",
				http.StatusInternalServerError,
			)

			return
		}
	}
}
