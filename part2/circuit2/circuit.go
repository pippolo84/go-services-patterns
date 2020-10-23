package circuit2

import (
	"context"
	"errors"
	"net/http"
	"time"

	// use:
	// - go mod init <module-name>
	// and
	// - go get -u github.com/cenkalti/backoff/v4@v4.1.0
	// to import the package
	"github.com/cenkalti/backoff/v4"
)

// Requester is an interface to wrap the Do method
type Requester interface {
	Do(req *http.Request) (*http.Response, error)
}

// ErrBreakerOpen is the error returned when the circuit breaker is open
var ErrBreakerOpen error = errors.New("circuit breaker is open")

// Client is a wrapper for a HTTP client that adds a circuit breaker
// to manage failures and mitigate the thundering herd problem
type Client struct {
	client *http.Client
	state  breakerState

	nfail     uint64
	threshold uint64

	timeout     time.Time
	cooldown    time.Duration
	openBackOff *backoff.ExponentialBackOff
}

type breakerState string

const (
	open     breakerState = "open"
	halfOpen breakerState = "half-open"
	closed   breakerState = "closed"
)

// NewDefaultClient returns the default http package client wrapped
// with an exponential backoff circuit breaker
// t is the threshold set to switch to the open state, expressed in
// number of consecutive failures before opening the circuit
func NewDefaultClient(t uint64) *Client {
	return &Client{
		client:      http.DefaultClient,
		state:       closed,
		threshold:   t,
		openBackOff: backoff.NewExponentialBackOff(),
	}
}

// NewClient returns the client passed as input wrapped
// with an exponential backoff circuit breaker
// t is the threshold set to switch to the open state, expressed in
// number of consecutive failures before opening the circuit
func NewClient(c *http.Client, t uint64) *Client {
	return &Client{
		client:      c,
		state:       closed,
		threshold:   t,
		openBackOff: backoff.NewExponentialBackOff(),
	}
}

// pre updates circuit breaker state before executing operation
// pre can prevent the operation to be executed returning an error
func (c *Client) pre() error {
	if c.state == open {
		if !time.Now().After(c.timeout) {
			return ErrBreakerOpen
		}

		// set to halfOpen and stay in that state for another cooldown interval
		c.state = halfOpen
		c.timeout = time.Now().Add(c.cooldown)
	}
	return nil
}

// post updates circuit breaker state after executing operation
func (c *Client) post(err error, statusCode int) {
	if c.state == open {
		panic("circuit breaker internal state corrupted")
	}

	if err == nil && statusCode < http.StatusInternalServerError ||
		err != nil && errors.Is(err, context.Canceled) {
		if c.state == halfOpen && time.Now().After(c.timeout) {
			// no errors since the circuitbreaker was set to half open
			c.state = closed
		}

		// reset consecutive failures counter
		c.nfail = 0

		return
	}

	c.nfail++
	switch c.state {
	case closed:
		if c.nfail >= c.threshold {
			c.state = open
			c.openBackOff.Reset()
			c.cooldown = c.openBackOff.NextBackOff()
			c.timeout = time.Now().Add(c.cooldown)
		}
	case halfOpen:
		c.state = open
		// set an exponential growing greater timeout
		if next := c.openBackOff.NextBackOff(); next != backoff.Stop {
			c.cooldown = next
		}
		c.timeout = time.Now().Add(c.cooldown)
	}
}

// Do sends the HTTP request req, returning its response.
// If the internal circuit breaker is open, it returns the
// ErrBreakerOpen error
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if err := c.pre(); err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)

	statusCode := 0
	if err == nil {
		statusCode = resp.StatusCode
	}

	c.post(err, statusCode)

	return resp, err
}
