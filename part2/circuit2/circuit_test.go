package circuit2

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBreakerSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, test!")
	}))

	client := NewDefaultClient(1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got: %d\n", http.StatusOK, resp.StatusCode)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	if string(buf) != "Hello, test!" {
		t.Fatalf("expected \"Hello, test!\", got: %q\n", string(buf))
	}

	// test that the circuit breaker is still closed
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got: %d\n", http.StatusOK, resp.StatusCode)
	}

	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	if string(buf) != "Hello, test!" {
		t.Fatalf("expected \"Hello, test!\", got: %q\n", string(buf))
	}
}

func TestBreakerNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	client := NewDefaultClient(1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, got: %d\n", http.StatusNotFound, resp.StatusCode)
	}

	// test that the circuit breaker is still closed
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, got: %d\n", http.StatusNotFound, resp.StatusCode)
	}
}

func TestBreakerServiceUnavailable(t *testing.T) {
	ts := httptest.NewServer(http.TimeoutHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// make the timeout handler trigger
			time.Sleep(20 * time.Millisecond)
		}),
		10*time.Millisecond,
		"timeout",
	))

	client := NewDefaultClient(1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status code %d, got: %d\n", http.StatusServiceUnavailable, resp.StatusCode)
	}

	// test that the circuit breaker is now open
	if _, err = client.Do(req); err != ErrBreakerOpen {
		t.Fatalf("expected ErrBreakerOpen error, got: %v\n", err)
	}
}

func TestBreakerOpenAndClose(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, test!")
	})
	mux.HandleFunc("/unavailable", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte("timeout")); err != nil {
			t.Fatal(err)
		}
	})
	ts := httptest.NewServer(mux)

	client := NewDefaultClient(1)

	successReq, err := http.NewRequest(http.MethodGet, ts.URL+"/ok", nil)
	if err != nil {
		t.Fatal(err)
	}

	failReq, err := http.NewRequest(http.MethodGet, ts.URL+"/unavailable", nil)
	if err != nil {
		t.Fatal(err)
	}

	// first request should fail and open the circuit breaker
	resp, err := client.Do(failReq)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status code %d, got: %d\n", http.StatusServiceUnavailable, resp.StatusCode)
	}

	if _, err = client.Do(successReq); err != ErrBreakerOpen {
		t.Fatalf("expected ErrBreakerOpen error, got: %v\n", err)
	}

	// wait to make the circuit breaker go to half open or closed state
	time.Sleep(time.Second)

	// the next request to /ok should be successful
	resp, err = client.Do(successReq)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got: %d\n", http.StatusOK, resp.StatusCode)
	}

	var buf []byte
	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	if string(buf) != "Hello, test!" {
		t.Fatalf("expected \"Hello, test!\", got: %q\n", string(buf))
	}
}

func TestBreakerFailureCounter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte("timeout")); err != nil {
			t.Fatal(err)
		}
	}))

	client := NewDefaultClient(10)

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		_, err = client.Do(req)
		if err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		}
	}

	// next request should find the circuit breaker in open state
	_, err = client.Do(req)
	if !errors.Is(err, ErrBreakerOpen) {
		t.Fatalf("expected ErrBreakerOpen error, got: %v\n", err)
	}
}

func TestBreakerFailureCounterReset(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, test!")
	})
	mux.HandleFunc("/unavailable", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte("timeout")); err != nil {
			t.Fatal(err)
		}
	})
	ts := httptest.NewServer(mux)

	client := NewDefaultClient(10)

	successReq, err := http.NewRequest(http.MethodGet, ts.URL+"/ok", nil)
	if err != nil {
		t.Fatal(err)
	}

	failReq, err := http.NewRequest(http.MethodGet, ts.URL+"/unavailable", nil)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 9; i++ {
		_, err = client.Do(failReq)
		if err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		}
	}

	// next successful request should reset the failure counter
	resp, err := client.Do(successReq)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got: %d\n", http.StatusOK, resp.StatusCode)
	}

	var buf []byte
	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	if string(buf) != "Hello, test!" {
		t.Fatalf("expected \"Hello, test!\", got: %q\n", string(buf))
	}

	// other 10 failed requests to open the circuit
	for i := 0; i < 10; i++ {
		_, err = client.Do(failReq)
		if err != nil {
			t.Fatalf("unexpected error: %v\n", err)
		}
	}

	// // this should fail
	_, err = client.Do(failReq)
	if !errors.Is(err, ErrBreakerOpen) {
		t.Fatalf("expected ErrBreakerOpen error, got: %v\n", err)
	}
}

func TestBreakerIncreasingInterval(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, test!")
	})
	mux.HandleFunc("/unavailable", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte("timeout")); err != nil {
			t.Fatal(err)
		}
	})
	ts := httptest.NewServer(mux)

	client := NewDefaultClient(1)

	successReq, err := http.NewRequest(http.MethodGet, ts.URL+"/ok", nil)
	if err != nil {
		t.Fatal(err)
	}

	failReq, err := http.NewRequest(http.MethodGet, ts.URL+"/unavailable", nil)
	if err != nil {
		t.Fatal(err)
	}

	var elapsed, nextElapsed time.Duration

	// first request should fail and open the circuit breaker
	start := time.Now()

	resp, err := client.Do(failReq)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status code %d, got: %d\n", http.StatusServiceUnavailable, resp.StatusCode)
	}

	// wait for the breaker to go to half open state
loop1:
	for {
		_, err = client.Do(successReq)
		switch err {
		case nil:
			elapsed = time.Since(start)
			break loop1
		case ErrBreakerOpen:
			time.Sleep(50 * time.Millisecond)
			continue
		default:
			t.Fatal(err)
		}
	}

	// next request should fail and set the circuit breaker to open state
	start = time.Now()

	resp, err = client.Do(failReq)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status code %d, got: %d\n", http.StatusServiceUnavailable, resp.StatusCode)
	}

	// wait for the breaker to go to half open state
loop2:
	for {
		_, err = client.Do(successReq)
		switch err {
		case nil:
			nextElapsed = time.Since(start)
			break loop2
		case ErrBreakerOpen:
			continue
		default:
			t.Fatal(err)
		}
	}

	// the elapsed time in open state should be greater the second time
	if nextElapsed <= elapsed {
		t.Fatalf("expected %v > %v", nextElapsed, elapsed)
	}
}

// run with: go test -v -run=Benchmark -bench=. -benchtime=10x
func BenchmarkBreakerIncreasingInterval(b *testing.B) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, test!")
	})
	mux.HandleFunc("/unavailable", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte("timeout")); err != nil {
			b.Fatal(err)
		}
	})
	ts := httptest.NewServer(mux)

	successReq, err := http.NewRequest(http.MethodGet, ts.URL+"/ok", nil)
	if err != nil {
		b.Fatal(err)
	}

	failReq, err := http.NewRequest(http.MethodGet, ts.URL+"/unavailable", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := NewDefaultClient(1)

		b.Logf("iteration %d of %d\n", i+1, b.N)

		for j := 0; j <= i; j++ {
			// first request should fail and open the circuit breaker
			start := time.Now()

			resp, err := client.Do(failReq)
			if err != nil {
				b.Fatalf("unexpected error: %v\n", err)
			}

			if resp.StatusCode != http.StatusServiceUnavailable {
				b.Fatalf("expected status code %d, got: %d\n", http.StatusServiceUnavailable, resp.StatusCode)
			}

			resp.Body.Close()

			// wait for the breaker to go to half open state
		loop1:
			for {
				_, err = client.Do(successReq)
				switch err {
				case nil:
					b.Log(time.Since(start))
					break loop1
				case ErrBreakerOpen:
					time.Sleep(5 * time.Millisecond)
					continue
				default:
					b.Fatal(err)
				}
			}
		}
	}

}
