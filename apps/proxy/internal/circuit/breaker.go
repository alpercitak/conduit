package circuit

import (
	"sync"
	"time"
)

type State string

const (
	Closed   State = "CLOSED"
	Open     State = "OPEN"
	HalfOpen State = "HALF_OPEN"
)

type Breaker struct {
	mu sync.Mutex

	state State

	failureCount int
	lastFailTime time.Time

	failureThreshold int
	retryTimeout     time.Duration

	// probeStart is when the in-flight half-open probe was handed out.
	// Zero means no probe is in flight.
	probeStart time.Time

	now func() time.Time
}

func New(failureThreshold int, retryTimeout time.Duration) *Breaker {
	return &Breaker{
		state:            Closed,
		failureThreshold: failureThreshold,
		retryTimeout:     retryTimeout,
		now:              time.Now,
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

	b.state = Closed
	b.failureCount = 0
	b.probeStart = time.Time{}
}

func (b *Breaker) Failure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failureCount++
	b.lastFailTime = b.now()
	b.probeStart = time.Time{}

	if b.state == HalfOpen || b.failureCount >= b.failureThreshold {
		b.state = Open
	}
}
