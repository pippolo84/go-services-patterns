package hedged

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Client is a wrapper for a HTTP client that implement
// the hedged requests pattern to lower tail latency
type Client struct {
	client      *http.Client
	delay       time.Duration
	concurrency int
}

// NewClient returns the client passed as input wrapped
// with a hedged requests logic
// The maximum concurrency level is set to n and
// the delay between each request is set to d
func NewClient(c *http.Client, d time.Duration, n int) *Client {
	return &Client{
		client:      c,
		delay:       d,
		concurrency: n,
	}
}

// Do sends the HTTP request req, returning the response and an error, if any
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	data := make(chan *http.Response, c.concurrency)
	errs := make(chan error, c.concurrency)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reqCtx := req.WithContext(ctx)

	n := 0
	for {
		if n < c.concurrency {
			n++
			go func() {
				res, err := c.client.Do(reqCtx)
				if err == nil {
					data <- res
					return
				}
				if !errors.Is(err, context.Canceled) {
					errs <- err
				}
			}()
		}

		select {
		case <-time.After(c.delay):
			continue
		case resp := <-data:
			return resp, nil
		case err := <-errs:
			return nil, err
		}
	}
}
