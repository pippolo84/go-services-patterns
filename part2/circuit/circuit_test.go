package circuit

import (
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

	breaker := NewBreaker(5*time.Second, 10)
	defer breaker.Stop()

	resp, err := breaker.Do(func() (*http.Response, error) {
		return http.Get(ts.URL)
	})
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
}

func TestBreakerNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	breaker := NewBreaker(5*time.Second, 10)
	defer breaker.Stop()

	resp, err := breaker.Do(func() (*http.Response, error) {
		return http.Get(ts.URL)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, got: %d\n", http.StatusNotFound, resp.StatusCode)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	if string(buf) != "404 page not found\n" {
		t.Fatalf("expected %q, got: %q\n", "404 page not found\n", string(buf))
	}
}

func TestBreakerServiceUnavailable(t *testing.T) {
	ts := httptest.NewServer(http.TimeoutHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(20 * time.Millisecond) }),
		10*time.Millisecond,
		"timeout",
	))

	breaker := NewBreaker(5*time.Second, 10)
	defer breaker.Stop()

	resp, err := breaker.Do(func() (*http.Response, error) {
		return http.Get(ts.URL)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status code %d, got: %d\n", http.StatusServiceUnavailable, resp.StatusCode)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	if string(buf) != "timeout" {
		t.Fatalf("expected %q, got: %q\n", "timeout", string(buf))
	}

	if breaker.nfail != 1 {
		t.Fatalf("expected breaker nfail 1, got: %q\n", breaker.nfail)
	}
}
