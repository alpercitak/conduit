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
}

func New(failureThreshold int, retryTimeout time.Duration) *Breaker {
	return &Breaker{
		state:            Closed,
		failureThreshold: failureThreshold,
		retryTimeout:     retryTimeout,
	}
}

func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {

	case Open:
		if time.Since(b.lastFailTime) > b.retryTimeout {
			b.state = HalfOpen
			return true
		}
		return false

	case HalfOpen:
		return true

	default:
		return true
	}
}

func (b *Breaker) Success() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.state = Closed
	b.failureCount = 0
}

func (b *Breaker) Failure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failureCount++
	b.lastFailTime = time.Now()

	if b.failureCount >= b.failureThreshold {
		b.state = Open
	}
}
