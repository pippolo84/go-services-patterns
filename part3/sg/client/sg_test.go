package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientOK(t *testing.T) {
	// uncomment this to see the goroutines leaking!
	//defer goleak.VerifyNone(t)

	testCases := []struct {
		name        string
		concurrency int
	}{
		{
			name:        "concurrent client n=1",
			concurrency: 1,
		},
		{
			name:        "concurrent client n=2",
			concurrency: 2,
		},
		{
			name:        "concurrent client n=25",
			concurrency: 25,
		},
		{
			name:        "concurrent client n=50",
			concurrency: 50,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClient(http.DefaultClient, tc.concurrency)
			req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
			if err != nil {
				t.Error(err)
			}

			resp, err := c.Do(req)
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("expecting status code %d, got: %v\n", http.StatusOK, resp.StatusCode)
			}
		})
	}
}

func TestClientNotFound(t *testing.T) {
	testCases := []struct {
		name        string
		concurrency int
	}{
		{
			name:        "concurrent client n=1",
			concurrency: 1,
		},
		{
			name:        "concurrent client n=2",
			concurrency: 2,
		},
		{
			name:        "concurrent client n=25",
			concurrency: 25,
		},
		{
			name:        "concurrent client n=50",
			concurrency: 50,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}))
	defer ts.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClient(http.DefaultClient, tc.concurrency)
			req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
			if err != nil {
				t.Error(err)
			}

			resp, err := c.Do(req)
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("expecting status code %d, got: %v\n", http.StatusNotFound, resp.StatusCode)
			}
		})
	}
}

func TestClientInternalError(t *testing.T) {
	testCases := []struct {
		name        string
		concurrency int
	}{
		{
			name:        "concurrent client n=1",
			concurrency: 1,
		},
		{
			name:        "concurrent client n=2",
			concurrency: 2,
		},
		{
			name:        "concurrent client n=25",
			concurrency: 25,
		},
		{
			name:        "concurrent client n=50",
			concurrency: 50,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}))
	defer ts.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClient(http.DefaultClient, tc.concurrency)
			req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
			if err != nil {
				t.Error(err)
			}

			resp, err := c.Do(req)
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusInternalServerError {
				t.Errorf("expecting status code %d, got: %v\n", http.StatusInternalServerError, resp.StatusCode)
			}
		})
	}
}
