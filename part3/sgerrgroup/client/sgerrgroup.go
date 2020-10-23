package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"golang.org/x/sync/errgroup"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reqCtx := req.WithContext(ctx)
	g, _ := errgroup.WithContext(ctx)
	for i := 0; i < c.concurrency; i++ {
		g.Go(func() error {
			resp, err := c.client.Do(reqCtx)
			if err != nil {
				return err
			}

			cancel()

			data <- resp

			return nil
		})
	}

	err := g.Wait()
	var e *url.Error
	if err != nil && !errors.As(err, &e) {
		return nil, err
	}

	return <-data, nil
}
