package tracer

import (
	"encoding/json"
	"net/http"

	"github.com/alpercitak/conduit/apps/proxy/internal/core"
)

func HTTPHandler(t *core.InMemoryTracer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("id")

		if id == "" {
			http.Error(w, "missing id", 400)
			return
		}

		trace, ok := t.Get(id)
		if !ok {
			http.Error(w, "trace not found", 404)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trace)
	}
}
