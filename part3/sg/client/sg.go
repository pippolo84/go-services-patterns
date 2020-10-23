package client

import (
	"net/http"
	"strings"
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
	data := make(chan *http.Response)
	errs := make(chan error)

	for i := 0; i < c.concurrency; i++ {
		go func() {
			resp, err := c.client.Do(req)
			if err == nil {
				data <- resp
				return
			}

			if !strings.Contains(err.Error(), "cancel") {
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
