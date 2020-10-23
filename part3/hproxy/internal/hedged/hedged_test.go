package hedged

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())

	os.Exit(m.Run())
}

func TestClientOK(t *testing.T) {
	defer goleak.VerifyNone(t)

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
		x := rand.Float32()
		switch {
		case x >= 0.999:
			time.Sleep(1000 * time.Millisecond)
		case x >= 0.99:
			time.Sleep(800 * time.Millisecond)
		case x >= 0.95:
			time.Sleep(80 * time.Millisecond)
		case x >= 0.5:
			time.Sleep(10 * time.Millisecond)
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}))
	// WARNING: forgetting to close the server will make goleak angry!
	defer ts.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClient(http.DefaultClient, 100*time.Millisecond, tc.concurrency)
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
	defer goleak.VerifyNone(t)

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
		x := rand.Float32()
		switch {
		case x >= 0.999:
			time.Sleep(1000 * time.Millisecond)
		case x >= 0.99:
			time.Sleep(800 * time.Millisecond)
		case x >= 0.95:
			time.Sleep(80 * time.Millisecond)
		case x >= 0.5:
			time.Sleep(10 * time.Millisecond)
		default:
			time.Sleep(5 * time.Millisecond)
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}))
	defer ts.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClient(http.DefaultClient, 100*time.Millisecond, tc.concurrency)
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
	defer goleak.VerifyNone(t)

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
		x := rand.Float32()
		switch {
		case x >= 0.999:
			time.Sleep(1000 * time.Millisecond)
		case x >= 0.99:
			time.Sleep(800 * time.Millisecond)
		case x >= 0.95:
			time.Sleep(80 * time.Millisecond)
		case x >= 0.5:
			time.Sleep(10 * time.Millisecond)
		default:
			time.Sleep(5 * time.Millisecond)
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}))
	defer ts.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClient(http.DefaultClient, 100*time.Millisecond, tc.concurrency)
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
