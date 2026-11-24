package circuit

import (
	"testing"
	"time"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time          { return c.t }
func (c *clock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newTestBreaker(threshold int, timeout time.Duration) (*Breaker, *clock) {
	c := &clock{t: time.Unix(0, 0)}
	b := New(threshold, timeout)
	b.now = c.now
	return b, c
}

func TestOpensAfterThreshold(t *testing.T) {
	b, _ := newTestBreaker(3, time.Second)

	b.Failure()
	b.Failure()
	if !b.Allow() {
		t.Fatal("should allow below threshold")
	}
	b.Failure()
	if b.State() != Open || b.Allow() {
		t.Fatalf("want OPEN and refusing, got %s", b.State())
	}
}

func TestAvailableDoesNotConsumeProbe(t *testing.T) {
	b, c := newTestBreaker(1, time.Second)
	b.Failure()
	c.advance(2 * time.Second)

	for range 5 {
		if !b.Available() {
			t.Fatal("Available should stay true until a probe is claimed")
		}
	}
	if b.State() != Open {
		t.Fatalf("Available must not change state, got %s", b.State())
	}
	if !b.Allow() {
		t.Fatal("first Allow after timeout should hand out the probe")
	}
	if b.Available() || b.Allow() {
		t.Fatal("only one probe while half-open")
	}
}

func TestProbeSuccessCloses(t *testing.T) {
	b, c := newTestBreaker(1, time.Second)
	b.Failure()
	c.advance(2 * time.Second)
	b.Allow()

	b.Success()
	if b.State() != Closed || !b.Allow() {
		t.Fatalf("want CLOSED, got %s", b.State())
	}
}

func TestProbeFailureReopens(t *testing.T) {
	b, c := newTestBreaker(3, time.Second)
	for range 3 {
		b.Failure()
	}
	c.advance(2 * time.Second)
	b.Allow()

	b.Failure()
	if b.State() != Open || b.Allow() {
		t.Fatalf("want OPEN after failed probe, got %s", b.State())
	}
}

func TestLostProbeIsReissued(t *testing.T) {
	b, c := newTestBreaker(1, time.Second)
	b.Failure()
	c.advance(2 * time.Second)
	b.Allow() // probe never reports back

	c.advance(2 * time.Second)
	if !b.Allow() {
		t.Fatal("a probe that never reported should be reissued after retryTimeout")
	}
}
