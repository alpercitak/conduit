package metrics

import (
	"fmt"
	"net/http"
)

func PrometheusHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p50, p95, p99, reqs, errs := store.Summary()

		w.Header().Set("Content-Type", "text/plain")

		fmt.Fprintf(w, "requests_total %d\n", reqs)
		fmt.Fprintf(w, "errors_total %d\n", errs)

		fmt.Fprintf(w, "latency_p50_ms %d\n", p50.Milliseconds())
		fmt.Fprintf(w, "latency_p95_ms %d\n", p95.Milliseconds())
		fmt.Fprintf(w, "latency_p99_ms %d\n", p99.Milliseconds())
	}
}
