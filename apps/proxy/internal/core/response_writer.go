package core

import "net/http"

type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int

	wroteHeader bool
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		StatusCode:     200,
	}
}

// WriteHeader records the first final status. 1xx informational headers
// (forwarded by ReverseProxy) pass through without fixing the status.
func (rw *ResponseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	if code >= 200 {
		rw.StatusCode = code
		rw.wroteHeader = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

// Write without a prior WriteHeader sends an implicit 200.
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	rw.wroteHeader = true
	return rw.ResponseWriter.Write(b)
}
