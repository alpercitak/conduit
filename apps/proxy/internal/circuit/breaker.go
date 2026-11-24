package circuit

import (
	"sync"
	"time"
)

type State string

// errorAlpha weights each result in the error-rate EWMA (~20-request memory).
// Tripping on a rate, not N failures in a row, keeps a steadily flaky
// upstream from opening just because it gets a lot of traffic.
const errorAlpha = 0.05

const (
	Closed   State = "CLOSED"
	Open     State = "OPEN"
	HalfOpen State = "HALF_OPEN"
)

type Breaker struct {
	mu sync.Mutex

	state State

	errorRate    float64 // EWMA, 0..1
	lastFailTime time.Time

	errorThreshold float64
	retryTimeout   time.Duration

	// probeStart is when the in-flight half-open probe was handed out.
	// Zero means no probe is in flight.
	probeStart time.Time

	now func() time.Time
}

// New returns a breaker that opens once the error-rate EWMA reaches
// errorThreshold (0..1) and retries after retryTimeout.
func New(errorThreshold float64, retryTimeout time.Duration) *Breaker {
	return &Breaker{
		state:          Closed,
		errorThreshold: errorThreshold,
		retryTimeout:   retryTimeout,
		now:            time.Now,
	}
}

// Available reports whether Allow would succeed right now, without
// changing state. Use it to filter candidates; call Allow only for the
// upstream that will actually receive the request.
func (b *Breaker) Available() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.canPass()
}

// Allow claims permission to send a request. In half-open state only one
// probe is handed out until Success or Failure is reported (or the probe
// times out, so a lost probe can't strand the upstream).
func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.canPass() {
		return false
	}

	if b.state == Open {
		b.state = HalfOpen
	}
	if b.state == HalfOpen {
		b.probeStart = b.now()
	}
	return true
}

func (b *Breaker) canPass() bool {
	switch b.state {
	case Open:
		return b.now().Sub(b.lastFailTime) > b.retryTimeout
	case HalfOpen:
		return b.probeStart.IsZero() || b.now().Sub(b.probeStart) > b.retryTimeout
	default:
		return true
	}
}

func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

func (b *Breaker) Success() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// A request sent before the breaker opened can still finish OK.
	// Only a half-open probe may close an open breaker.
	if b.state == Open {
		return
	}

	if b.state == HalfOpen {
		b.state = Closed
		b.errorRate = 0 // start fresh, or the old rate re-trips it at once
	} else {
		b.errorRate *= 1 - errorAlpha
	}
	b.probeStart = time.Time{}
}

func (b *Breaker) Failure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lastFailTime = b.now()
	b.probeStart = time.Time{}

	switch b.state {
	case HalfOpen:
		b.state = Open
	case Closed:
		b.errorRate = errorAlpha + (1-errorAlpha)*b.errorRate
		if b.errorRate >= b.errorThreshold {
			b.state = Open
		}
	}
}
