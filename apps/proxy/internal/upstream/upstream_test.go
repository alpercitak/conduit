package upstream

import (
	"math"
	"sync"
	"testing"
	"time"
)

func TestSingleFailureAfterSuccessesIsNotFatal(t *testing.T) {
	u, _ := New("a", "http://localhost", nil)
	for range 10 {
		u.Record(time.Millisecond, false)
	}
	u.Record(time.Millisecond, true)

	_, errRate := u.Stats()
	if math.Abs(errRate-alphaError) > 1e-9 {
		t.Fatalf("error rate = %v, want %v", errRate, alphaError)
	}
}

func TestSubMillisecondLatencyIsKept(t *testing.T) {
	u, _ := New("a", "http://localhost", nil)
	u.Record(300*time.Microsecond, false)
	u.Record(300*time.Microsecond, false)

	lat, _ := u.Stats()
	if math.Abs(lat-0.3) > 1e-9 {
		t.Fatalf("latency = %v ms, want 0.3", lat)
	}
}

func TestLatencyEWMA(t *testing.T) {
	u, _ := New("a", "http://localhost", nil)
	u.Record(100*time.Millisecond, false)
	u.Record(200*time.Millisecond, false)

	lat, _ := u.Stats()
	want := alphaLatency*200 + (1-alphaLatency)*100
	if math.Abs(lat-want) > 1e-9 {
		t.Fatalf("latency = %v, want %v", lat, want)
	}
}

func TestConcurrentRecordAndStats(t *testing.T) {
	u, _ := New("a", "http://localhost", nil)
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(2)
		go func() { defer wg.Done(); u.Record(time.Millisecond, i%2 == 0) }()
		go func() { defer wg.Done(); u.Stats() }()
	}
	wg.Wait()
}
