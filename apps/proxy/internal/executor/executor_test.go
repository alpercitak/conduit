package executor

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"proxy/internal/circuit"
	"proxy/internal/core"
	"proxy/internal/filter"
	"proxy/internal/metrics"
	"proxy/internal/router"
	"proxy/internal/upstream"
)

// metrics.New registers on the global Prometheus registry, so it can run once.
var testMetrics = metrics.New()

// newExecutor wires a single upstream the same way cmd/proxy does.
func newExecutor(t *testing.T, url string) (*Executor, *upstream.Upstream) {
	t.Helper()
	u, err := upstream.New("alpha", url, circuit.New(0.01, time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	pool := upstream.NewPool([]*upstream.Upstream{u})
	chain := filter.NewChain(filter.NewResilienceFilter(func(ctx *core.Context) *upstream.Upstream {
		return pool.GetByID(ctx.UpstreamID)
	}))
	return New(router.NewSmart(pool), core.NewInMemoryTracer(), chain, testMetrics), u
}

func serve(e *Executor) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.Handle(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec
}

func TestProxiesAndRecordsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "ok")
	}))
	defer srv.Close()

	e, u := newExecutor(t, srv.URL)
	rec := serve(e)

	if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
		t.Fatalf("got %d %q", rec.Code, rec.Body.String())
	}
	if latency, errRate := u.Stats(); latency <= 0 || errRate != 0 {
		t.Fatalf("want latency recorded and no errors, got %v / %v", latency, errRate)
	}
	if u.Breaker.State() != circuit.Closed {
		t.Fatalf("want CLOSED, got %s", u.Breaker.State())
	}
}

func TestUpstream5xxOpensBreakerAndStopsRouting(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	e, u := newExecutor(t, srv.URL)

	if rec := serve(e); rec.Code != http.StatusInternalServerError {
		t.Fatalf("first request: want 500, got %d", rec.Code)
	}
	if _, errRate := u.Stats(); errRate <= 0 {
		t.Fatal("want failure recorded in error rate")
	}
	if u.Breaker.State() != circuit.Open {
		t.Fatalf("want OPEN, got %s", u.Breaker.State())
	}
	before := testutil.ToFloat64(testMetrics.Errors.WithLabelValues(noUpstream))
	if rec := serve(e); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("second request: want 503, got %d", rec.Code)
	}
	if got := testutil.ToFloat64(testMetrics.Errors.WithLabelValues(noUpstream)); got != before+1 {
		t.Fatalf("want rejected request counted as error, got %v -> %v", before, got)
	}
}

func TestUnreachableUpstreamCountsAsFailure(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()

	e, u := newExecutor(t, url)

	if rec := serve(e); rec.Code != http.StatusBadGateway {
		t.Fatalf("want 502, got %d", rec.Code)
	}
	if u.Breaker.State() != circuit.Open {
		t.Fatalf("want OPEN, got %s", u.Breaker.State())
	}
}

func TestClientCancelDoesNotCountAgainstUpstream(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()

	e, u := newExecutor(t, srv.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rec := httptest.NewRecorder()
	e.Handle(rec, httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx))

	if rec.Code != upstream.StatusClientClosedRequest {
		t.Fatalf("want 499, got %d", rec.Code)
	}
	if u.Breaker.State() != circuit.Closed {
		t.Fatalf("want CLOSED, got %s", u.Breaker.State())
	}
	if _, errRate := u.Stats(); errRate != 0 {
		t.Fatalf("want no error recorded, got %v", errRate)
	}
}
