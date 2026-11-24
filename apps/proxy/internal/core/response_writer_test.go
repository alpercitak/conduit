package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFirstFinalStatusWins(t *testing.T) {
	cases := []struct {
		name  string
		write func(rw *ResponseWriter)
		want  int
	}{
		{"explicit 200 then 500", func(rw *ResponseWriter) {
			rw.WriteHeader(http.StatusOK)
			rw.WriteHeader(http.StatusInternalServerError)
		}, http.StatusOK},
		{"implicit 200 then 500", func(rw *ResponseWriter) {
			rw.Write([]byte("x"))
			rw.WriteHeader(http.StatusInternalServerError)
		}, http.StatusOK},
		{"1xx then 502", func(rw *ResponseWriter) {
			rw.WriteHeader(http.StatusEarlyHints)
			rw.WriteHeader(http.StatusBadGateway)
		}, http.StatusBadGateway},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			rw := NewResponseWriter(rec)
			c.write(rw)

			if rw.StatusCode != c.want {
				t.Fatalf("recorded %d, want %d", rw.StatusCode, c.want)
			}
			// ResponseRecorder keeps a 1xx as its Code, so only check what
			// was sent when no informational header went first.
			if c.want == http.StatusOK && rec.Code != c.want {
				t.Fatalf("sent %d, want %d", rec.Code, c.want)
			}
		})
	}
}

func TestFlushReachesUnderlyingWriter(t *testing.T) {
	rec := httptest.NewRecorder()

	if err := http.NewResponseController(NewResponseWriter(rec)).Flush(); err != nil {
		t.Fatal(err)
	}
	if !rec.Flushed {
		t.Fatal("want underlying writer flushed")
	}
}
