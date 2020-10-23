package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pippolo84/go-services-patterns/part2/threader/internal/model"
)

func TestFailures(t *testing.T) {
	a := Archive{
		store: model.NewAnalyticsStore(),
	}
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(a.newScore))

	restore := a.restore

	// wait for a failure
	for {
		resp, err := http.Get(ts.URL)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode == http.StatusUnsupportedMediaType {
			resp.Body.Close()
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatal(err)
		}

		break
	}

	if a.nfail != 1 {
		t.Fatalf("expected failures count = 1, got: %d\n", a.nfail)
	}

	if a.restore == restore {
		t.Fatal("expected restore time to be updated, but remained the same")
	}

	// next request should fail and increase restore deadline
	restore = a.restore
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status code %d, got: %d\n", http.StatusInternalServerError, resp.StatusCode)
	}

	if a.nfail != 2 {
		t.Fatalf("expected failures count = 2, got: %d\n", a.nfail)
	}

	if a.restore == restore {
		t.Fatal("expected restore time to be updated, but remained the same")
	}
}
