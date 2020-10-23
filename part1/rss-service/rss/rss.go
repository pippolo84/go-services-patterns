package rss

import (
	"context"
	"net/http"

	"github.com/mmcdole/gofeed"
)

// Fetcher is an interface that encapsulate the Fetch method
type Fetcher interface {
	// FetchWithContext retrieves all the RSS items from a RSS specified by its URL,
	// returning an error if something goes wrong
	// It accepts a context to support cancellation
	FetchWithContext(ctx context.Context, url string) ([]*gofeed.Item, error)

	// Fetch retrieves all the RSS items from a RSS specified by its URL,
	// returning an error if something goes wrong
	Fetch(url string) ([]*gofeed.Item, error)
}

// Client is a type implement the RSSSourcer interface
type Client struct {
	// Parser is a reference to the gofeed Parser used
	// to parse received RSS items
	Parser *gofeed.Parser
}

// NewClient returns a new Client
func NewClient() *Client {
	return &Client{gofeed.NewParser()}
}

// FetchWithContext retrieves all the RSS items from a RSS specified by its URL,
// returning an error if something goes wrong
// It accepts a context to support cancellation
func (rc *Client) FetchWithContext(ctx context.Context, url string) ([]*gofeed.Item, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	feed, err := rc.Parser.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return feed.Items, nil
}

// Fetch retrieves all the RSS items from a RSS specified by its URL,
// returning an error if something goes wrong
func (rc *Client) Fetch(url string) ([]*gofeed.Item, error) {
	feed, err := rc.Parser.ParseURL(url)
	if err != nil {
		return nil, err
	}

	return feed.Items, nil
}
