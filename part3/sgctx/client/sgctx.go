package client

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"
)

// Client is a wrapper for a HTTP client that adds scatter-gather
// logic to lower tail latency
type Client struct {
	client      *http.Client
	concurrency int
}

// NewClient returns the client passed as input wrapped
// with a scatter-gather logic
// The maximum concurrency level is set to n
func NewClient(c *http.Client, n int) *Client {
	return &Client{
		client:      c,
		concurrency: n,
	}
}

// Do sends the HTTP request req, returning the response and an error, if any
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	data := make(chan *http.Response, c.concurrency)
	errs := make(chan error, c.concurrency)

	ctx, cancel := context.WithCancel(context.Background())
	reqCtx := req.WithContext(ctx)
	for i := 0; i < c.concurrency; i++ {
		go func() {
			defer cancel()

			resp, err := c.client.Do(reqCtx)
			if err == nil {
				data <- resp
				return
			}
			if !errors.Is(err, context.Canceled) {
				errs <- err
			}
		}()
	}

	select {
	case resp := <-data:
		return resp, nil
	case err := <-errs:
		return nil, err
	}
}
