package router

import (
	"testing"
	"time"

	"proxy/internal/circuit"
	"proxy/internal/upstream"
)

func newUpstream(t *testing.T, id string, b *circuit.Breaker) *upstream.Upstream {
	t.Helper()
	u, err := upstream.New(id, "http://localhost", b)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func newRouter(ups ...*upstream.Upstream) *SmartRouter {
	r := NewSmart(upstream.NewPool(ups))
	r.exploration = 0
	return r
}

func TestPicksLowerLatency(t *testing.T) {
	a := newUpstream(t, "alpha", nil)
	b := newUpstream(t, "beta", nil)
	a.Record(100*time.Millisecond, false)
	b.Record(10*time.Millisecond, false)

	got, err := newRouter(a, b).Pick()
	if err != nil || got != b {
		t.Fatalf("want beta, got %v (%v)", got, err)
	}
}

func TestPenalizesErrors(t *testing.T) {
	a := newUpstream(t, "alpha", nil)
	b := newUpstream(t, "beta", nil)
	a.Record(10*time.Millisecond, true)
	b.Record(50*time.Millisecond, false)

	got, _ := newRouter(a, b).Pick()
	if got != b {
		t.Fatalf("want beta, got %s", got.ID)
	}
}

func TestVerySlowUpstreamsStillPicked(t *testing.T) {
	a := newUpstream(t, "alpha", nil)
	b := newUpstream(t, "beta", nil)
	a.Record(3*time.Second, false)
	b.Record(2*time.Second, false)

	got, err := newRouter(a, b).Pick()
	if err != nil || got != b {
		t.Fatalf("want beta, got %v (%v)", got, err)
	}
}

func TestSkipsOpenBreaker(t *testing.T) {
	br := circuit.New(1, time.Hour)
	br.Failure()
	a := newUpstream(t, "alpha", br)
	b := newUpstream(t, "beta", nil)
	b.Record(500*time.Millisecond, false) // worse score, but alpha is open

	got, _ := newRouter(a, b).Pick()
	if got != b {
		t.Fatalf("want beta, got %s", got.ID)
	}
}

func TestNoneAvailable(t *testing.T) {
	br := circuit.New(1, time.Hour)
	br.Failure()

	if _, err := newRouter(newUpstream(t, "alpha", br)).Pick(); err == nil {
		t.Fatal("want error when every breaker is open")
	}
}

// Regression: checking a half-open upstream that isn't picked must not use
// up its probe, otherwise it is excluded forever.
func TestUnpickedRecoveringUpstreamIsNotStranded(t *testing.T) {
	br := circuit.New(1, 10*time.Millisecond)
	br.Failure()
	a := newUpstream(t, "alpha", nil)
	b := newUpstream(t, "beta", br)
	a.Record(time.Millisecond, false)
	b.Record(900*time.Millisecond, false)
	time.Sleep(20 * time.Millisecond)

	r := newRouter(a, b)
	for range 10 {
		if got, _ := r.Pick(); got != a {
			t.Fatalf("want alpha, got %s", got.ID)
		}
	}

	if !br.Available() {
		t.Fatal("beta's probe was consumed without beta being picked")
	}

	// Once alpha goes away, beta gets its probe.
	r = newRouter(b)
	if got, err := r.Pick(); err != nil || got != b {
		t.Fatalf("want beta probe, got %v (%v)", got, err)
	}
}
