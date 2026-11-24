package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"proxy/internal/circuit"
	"proxy/internal/upstream"
)

func TestUpstreamGaugesReadLiveState(t *testing.T) {
	b := circuit.New(0.01, time.Hour)
	u, err := upstream.New("alpha", "http://localhost", b)
	if err != nil {
		t.Fatal(err)
	}
	g := upstreamGauges(u, func(*upstream.Upstream) float64 { return 42 })
	latency, errRate, score, state := g[0], g[1], g[2], g[3]

	u.Record(20*time.Millisecond, true)
	b.Failure()

	if got := testutil.ToFloat64(latency); got != 20 {
		t.Fatalf("latency: want 20, got %v", got)
	}
	if got := testutil.ToFloat64(errRate); got <= 0 {
		t.Fatalf("error rate: want > 0, got %v", got)
	}
	if got := testutil.ToFloat64(score); got != 42 {
		t.Fatalf("score: want 42, got %v", got)
	}
	if got := testutil.ToFloat64(state); got != 2 {
		t.Fatalf("breaker: want 2 (open), got %v", got)
	}
}
