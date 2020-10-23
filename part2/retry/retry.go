package retry

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Strategy is an interface that wraps the BackOff method
type Strategy interface {
	BackOff(n int) time.Duration
}

// Constant implements the constant backoff algorithm
type Constant struct {
	// Quantum is the basic unit of duration for the algorithm
	Quantum time.Duration
}

// BackOff satisfies the Strategy interface for the Constant type
// it gets the current retry number to return the associated
// constant backoff duration
func (c Constant) BackOff(_ int) time.Duration {
	return c.Quantum
}

// JitteredConstant implements the constant backoff algorithm
// with a +/- 33% jitter
type JitteredConstant struct {
	Constant
}

// BackOff satisfies the Strategy interface for the JitteredConstant type
// it gets the current retry number to return the associated
// jittered constant backoff duration
func (jc JitteredConstant) BackOff(n int) time.Duration {
	return jitter(jc.Constant.BackOff(n))
}

// Linear implements the linear backoff algorithm
type Linear struct {
	// Quantum is the basic unit of duration for the algorithm
	Quantum time.Duration
}

// BackOff satisfies the Strategy interface for the Linear type
// it gets the current retry number to return the associated
// linear backoff duration
func (l Linear) BackOff(n int) time.Duration {
	return time.Duration(n) * l.Quantum
}

// JitteredLinear implements the linear backoff algorithm
// with a +/- 33% jitter
type JitteredLinear struct {
	Linear
}

// BackOff satisfies the Strategy interface for the JitteredLinear type
// it gets the current retry number to return the associated
// jittered linear backoff duration
func (jl JitteredLinear) BackOff(n int) time.Duration {
	return jitter(jl.Linear.BackOff(n))
}

// Exponential implements the exponential backoff algorithm
type Exponential struct {
	// Quantum is the basic unit of duration for the algorithm
	Quantum time.Duration
}

// BackOff satisfies the Strategy interface for the Exponential type
// it gets the current retry number to return the associated
// exponential backoff duration
func (e Exponential) BackOff(n int) time.Duration {
	return time.Duration(1<<n) * e.Quantum
}

// JitteredExponential implements the exponential backoff algorithm
// with a +/- 33% jitter
type JitteredExponential struct {
	Exponential
}

// BackOff satisfies the Strategy interface for the Exponential type
// it gets the current retry number to return the associated
// exponential backoff duration
func (je JitteredExponential) BackOff(n int) time.Duration {
	return jitter(je.Exponential.BackOff(n))
}

func jitter(d time.Duration) time.Duration {
	maxJitter := int64(d) / 3

	// add 1 to avoid rand.Int63n(0) case
	jitter := rand.Int63n(2*maxJitter+1) - maxJitter

	return d + time.Duration(jitter)
}

// Client is a wrapper for a HTTP client that adds a
// retry logic to manage temporary failures
type Client struct {
	client     *http.Client
	maxRetries int
	strategy   Strategy
}

// NewDefaultClient returns the default http package client wrapped
// with a retry logic
// strategy is the Strategy used to get a backoff time before the next retry
// maxRetries is the maximum number of retries before returning a failure
func NewDefaultClient(strategy Strategy, maxRetries int) *Client {
	return &Client{
		client:     http.DefaultClient,
		maxRetries: maxRetries,
		strategy:   strategy,
	}
}

// NewClient returns the client passed as input wrapped
// with a retry logic
// strategy is the Strategy used to get a backoff time before the next retry
// maxRetries is the maximum number of retries before returning a failure
func NewClient(c *http.Client, strategy Strategy, maxRetries int) *Client {
	return &Client{
		client:     c,
		maxRetries: maxRetries,
		strategy:   strategy,
	}
}

// Do sends the HTTP request req, returning its response.
// If the request fail, it waits a backoff time taken from the
// strategy set and try again
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		body []byte
		err  error
	)

	if req.Body != nil {
		body, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < c.maxRetries; i++ {
		// ugly hack to "rewind" the request body
		if req.Body != nil {
			req.Body = ioutil.NopCloser(bytes.NewReader(body))
		}
		resp, err = c.client.Do(req)
		if err == nil && resp.StatusCode < http.StatusInternalServerError ||
			err != nil && errors.Is(err, context.Canceled) {
			break
		}

		time.Sleep(c.strategy.BackOff(i))
	}

	return resp, err
}
