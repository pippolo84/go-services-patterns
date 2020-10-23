package circuit

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// FIXME: use expvar
// FIXME: use backoff in halfOpen state: https://pkg.go.dev/github.com/cenkalti/backoff/v4

// Breaker is a type that implements the circuit breaker pattern
type Breaker struct {
	state     breakerState
	nfail     uint64
	threshold uint64
	done      chan struct{}
}

type breakerState string

const (
	open     breakerState = "open"
	halfOpen breakerState = "half-open"
	closed   breakerState = "closed"
)

// NewBreaker returns a new Breaker with the specified
// tick duration and failures threshold number
func NewBreaker(tick time.Duration, threshold uint64) *Breaker {
	breaker := &Breaker{
		state:     closed,
		nfail:     0,
		threshold: threshold,
		done:      make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(tick)
		defer ticker.Stop()

		for {
			select {
			case <-breaker.done:
				return
			case <-ticker.C:
				switch breaker.state {
				case open:
					breaker.state = halfOpen
				case halfOpen:
					if breaker.nfail == 0 {
						breaker.state = closed
					} else {
						breaker.state = open
					}
				case closed:
					if breaker.nfail >= breaker.threshold {
						breaker.state = open
					}
				}
			}
		}
	}()

	return breaker
}

// Stop will stop the supporting goroutine for the breaker
func (b *Breaker) Stop() {
	close(b.done)
}

// ErrCircuitBreakerOpen is the error that the breaker returns when it is open
var ErrCircuitBreakerOpen error = errors.New("circuit breaker is open")

// Do checks the state of the circuit breaker and
// executes the task accordingly
func (b *Breaker) Do(task func() (*http.Response, error)) (*http.Response, error) {
	if b.state == open {
		return nil, ErrCircuitBreakerOpen
	}

	resp, err := task()

	if (err != nil && !errors.Is(err, context.Canceled)) || resp.StatusCode >= 500 {
		b.nfail++
	}

	return resp, err
}
