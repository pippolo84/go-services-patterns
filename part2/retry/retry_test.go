package retry

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	testCases := []struct {
		name      string
		algorithm Strategy
		retries   map[int]time.Duration
	}{
		{
			name:      "Constant Backoff",
			algorithm: Constant{time.Second},
			retries: map[int]time.Duration{
				0:  time.Second,
				1:  time.Second,
				2:  time.Second,
				3:  time.Second,
				4:  time.Second,
				5:  time.Second,
				10: time.Second,
				20: time.Second,
			},
		},
		{
			name:      "Linear Backoff",
			algorithm: Linear{time.Second},
			retries: map[int]time.Duration{
				0:  0,
				1:  time.Second,
				2:  2 * time.Second,
				3:  3 * time.Second,
				4:  4 * time.Second,
				5:  5 * time.Second,
				10: 10 * time.Second,
				20: 20 * time.Second,
			},
		},
		{
			name:      "Exponential Backoff",
			algorithm: Exponential{time.Second},
			retries: map[int]time.Duration{
				0:  time.Second,
				1:  2 * time.Second,
				2:  4 * time.Second,
				3:  8 * time.Second,
				4:  16 * time.Second,
				5:  32 * time.Second,
				10: 1024 * time.Second,
				20: 1024 * 1024 * time.Second,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for n, backoff := range tc.retries {
				got := tc.algorithm.BackOff(n)
				if got != backoff {
					t.Errorf("expected bac1333 * time.Millisecondkoff %v for retry %d, got: %v", backoff, n, got)
				}
			}
		})
	}
}

func TestJitteredBackoff(t *testing.T) {
	type interval struct {
		min time.Duration
		max time.Duration
	}

	testCases := []struct {
		name      string
		algorithm Strategy
		retries   map[int]interval
	}{
		{
			name:      "Constant Backoff",
			algorithm: JitteredConstant{Constant{time.Second}},
			retries: map[int]interval{
				0: {
					min: 666 * time.Millisecond,
					max: 1333 * time.Millisecond,
				},
				1: {
					min: 666 * time.Millisecond,
					max: 1333 * time.Millisecond,
				},
				2: {
					min: 666 * time.Millisecond,
					max: 1333 * time.Millisecond,
				},
				3: {
					min: 666 * time.Millisecond,
					max: 1333 * time.Millisecond,
				},
				4: {
					min: 666 * time.Millisecond,
					max: 1333 * time.Millisecond,
				},
				5: {
					min: 666 * time.Millisecond,
					max: 1333 * time.Millisecond,
				},
				10: {
					min: 666 * time.Millisecond,
					max: 1333 * time.Millisecond,
				},
				20: {
					min: 666 * time.Millisecond,
					max: 1333 * time.Millisecond,
				},
			},
		},
		{
			name:      "Linear Backoff",
			algorithm: JitteredLinear{Linear{time.Second}},
			retries: map[int]interval{
				0: {
					min: 0,
					max: 0,
				},
				1: {
					min: 666 * time.Millisecond,
					max: 1333 * time.Millisecond,
				},
				2: {
					min: 1333 * time.Millisecond,
					max: 2666 * time.Millisecond,
				},
				3: {
					min: 2000 * time.Millisecond,
					max: 4000 * time.Millisecond,
				},
				4: {
					min: 2666 * time.Millisecond,
					max: 5333 * time.Millisecond,
				},
				5: {
					min: 3333 * time.Millisecond,
					max: 6666 * time.Millisecond,
				},
				10: {
					min: 6666 * time.Millisecond,
					max: 13333 * time.Millisecond,
				},
				20: {
					min: 13333 * time.Millisecond,
					max: 26666 * time.Millisecond,
				},
			},
		},
		{
			name:      "Exponential Backoff",
			algorithm: JitteredExponential{Exponential{time.Second}},
			retries: map[int]interval{
				0: {
					min: 666 * time.Millisecond,
					max: 1334 * time.Millisecond,
				},
				1: {
					min: 1333 * time.Millisecond,
					max: 2667 * time.Millisecond,
				},
				2: {
					min: 2666 * time.Millisecond,
					max: 5334 * time.Millisecond,
				},
				3: {
					min: 5333 * time.Millisecond,
					max: 10667 * time.Millisecond,
				},
				4: {
					min: 10666 * time.Millisecond,
					max: 21334 * time.Millisecond,
				},
				5: {
					min: 21333 * time.Millisecond,
					max: 42667 * time.Millisecond,
				},
				10: {
					min: 682666 * time.Millisecond,
					max: 1365334 * time.Millisecond,
				},
				20: {
					min: 699050666 * time.Millisecond,
					max: 1398101334 * time.Millisecond,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for n, backoff := range tc.retries {
				got := tc.algorithm.BackOff(n)
				if got < backoff.min || got > backoff.max {
					t.Errorf(
						"expected backoff in [%v,%v] for retry %d, got: %v",
						backoff.min, backoff.max, n, got)
				}
			}
		})
	}
}

func TestRetryClientSuccess(t *testing.T) {
	testCases := []struct {
		name       string
		maxRetries int
		strategy   Strategy
	}{
		{
			name:       "Retry with constant backoff",
			maxRetries: 3,
			strategy:   Constant{5 * time.Millisecond},
		},
		{
			name:       "Retry with linear backoff",
			maxRetries: 3,
			strategy:   Linear{5 * time.Millisecond},
		},
		{
			name:       "Retry with exponential backoff",
			maxRetries: 3,
			strategy:   Exponential{5 * time.Millisecond},
		},
	}

	for _, tc := range testCases {
		n := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			n++
			w.WriteHeader(http.StatusOK)
		}))

		client := NewDefaultClient(tc.strategy, tc.maxRetries)

		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d, got: %d\n", http.StatusOK, resp.StatusCode)
		}

		if n != 1 {
			t.Fatalf("expected %s to be queried one time, got: %d\n", ts.URL, n)
		}
	}
}

func TestRetryClientFailure(t *testing.T) {
	testCases := []struct {
		name       string
		maxRetries int
		strategy   Strategy
	}{
		{
			name:       "Retry with constant backoff",
			maxRetries: 3,
			strategy:   Constant{5 * time.Millisecond},
		},
		{
			name:       "Retry with linear backoff",
			maxRetries: 3,
			strategy:   Linear{5 * time.Millisecond},
		},
		{
			name:       "Retry with exponential backoff",
			maxRetries: 3,
			strategy:   Exponential{5 * time.Millisecond},
		},
	}

	for _, tc := range testCases {
		n := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			n++
			w.WriteHeader(http.StatusServiceUnavailable)
			if _, err := w.Write([]byte("timeout")); err != nil {
				t.Fatal(err)
			}
		}))

		client := NewDefaultClient(tc.strategy, tc.maxRetries)

		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		}

		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("expected status code %d, got: %d\n", http.StatusServiceUnavailable, resp.StatusCode)
		}

		if n != tc.maxRetries {
			t.Fatalf("expected %s to be queried %d times, got: %d\n", ts.URL, tc.maxRetries, n)
		}
	}
}

func TestRequestRetry(t *testing.T) {
	testCases := []struct {
		name       string
		maxRetries int
		strategy   Strategy
	}{
		{
			name:       "Retry with constant backoff",
			maxRetries: 3,
			strategy:   Constant{5 * time.Millisecond},
		},
		{
			name:       "Retry with linear backoff",
			maxRetries: 3,
			strategy:   Linear{5 * time.Millisecond},
		},
		{
			name:       "Retry with exponential backoff",
			maxRetries: 3,
			strategy:   Exponential{5 * time.Millisecond},
		},
	}

	for _, tc := range testCases {
		n := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			n++
			if n >= 3 {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusServiceUnavailable)
		}))

		client := NewDefaultClient(tc.strategy, tc.maxRetries)

		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d, got: %d\n", http.StatusOK, resp.StatusCode)
		}

		if n != tc.maxRetries {
			t.Fatalf("expected %s to be queried 3 times, got: %d\n", ts.URL, n)
		}
	}
}
