package filter

import (
	"testing"
	"time"

	"proxy/internal/circuit"
	"proxy/internal/core"
	"proxy/internal/upstream"
)

func newResilience(t *testing.T) (*ResilienceFilter, *circuit.Breaker) {
	t.Helper()
	b := circuit.New(1, time.Hour)
	u, err := upstream.New("alpha", "http://localhost", b)
	if err != nil {
		t.Fatal(err)
	}
	return NewResilienceFilter(func(*core.Context) *upstream.Upstream { return u }), b
}

func TestResilienceOpensOn5xx(t *testing.T) {
	f, b := newResilience(t)

	_ = f.After(&core.Context{StatusCode: 502})

	if b.State() != circuit.Open {
		t.Fatalf("want OPEN, got %s", b.State())
	}
}

func TestResilienceClosesOnNon5xx(t *testing.T) {
	for _, code := range []int{200, 404} {
		f, b := newResilience(t)
		b.Failure()

		_ = f.After(&core.Context{StatusCode: code})

		if b.State() != circuit.Closed {
			t.Fatalf("status %d: want CLOSED, got %s", code, b.State())
		}
	}
}

func TestResilienceIgnoresUnknownUpstream(t *testing.T) {
	f := NewResilienceFilter(func(*core.Context) *upstream.Upstream { return nil })

	if err := f.After(&core.Context{StatusCode: 500}); err != nil {
		t.Fatal(err)
	}
}

func TestResilienceIgnoresClientCancel(t *testing.T) {
	f, b := newResilience(t)
	b.Failure()

	_ = f.After(&core.Context{StatusCode: upstream.StatusClientClosedRequest})

	if b.State() != circuit.Open {
		t.Fatalf("want OPEN unchanged, got %s", b.State())
	}
}
