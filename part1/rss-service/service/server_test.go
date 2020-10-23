// +build !integration

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

// test with `go test -v -race -timeout=30s .`

func TestFeedsHandler(t *testing.T) {
	testCases := []struct {
		name  string
		feeds map[string]string
	}{
		{
			name:  "no feeds",
			feeds: map[string]string{},
		},
		{
			name: "single feed",
			feeds: map[string]string{
				"test": "test-url",
			},
		},
		{
			name: "multiple feeds",
			feeds: map[string]string{
				"test1": "test-url1",
				"test2": "test-url2",
				"test3": "test-url3",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/feeds", nil)
			if err != nil {
				t.Fatal(err)
			}

			f, err := os.Open(os.DevNull)
			if err != nil {
				t.Fatal(err)
			}

			svc := Service{
				log:   log.New(f, "", log.LstdFlags),
				feeds: tc.feeds,
			}

			rr := httptest.NewRecorder()

			svc.getFeeds(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected status code %d, got %d\n", http.StatusOK, rr.Code)
			}

			if rr.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("expected Content-Type %s, got %s\n", "application/json", rr.Header().Get("Content-Type"))
			}

			var feeds []Feed
			if err := json.NewDecoder(rr.Body).Decode(&feeds); err != nil {
				t.Fatalf("unexpected error while decoding JSON response body: %v\n", err)
			}

			if len(feeds) != len(tc.feeds) {
				t.Fatalf("expected %d feeds in response, got %d\n\n", len(tc.feeds), len(feeds))
			}

			f.Close()
		})
	}
}

func TestFeedHandlerContentType(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	svc := Service{
		log:   log.New(f, "", log.LstdFlags),
		feeds: map[string]string{},
	}

	req, err := http.NewRequest(http.MethodPost, "/feed", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "invalid")
	rr := httptest.NewRecorder()

	svc.addFeed(rr, req)

	if rr.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status code %d, got %d\n", http.StatusUnsupportedMediaType, rr.Code)
	}
}

func TestFeedHandlerBadlyFormedJSON(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	svc := Service{
		log:   log.New(f, "", log.LstdFlags),
		feeds: map[string]string{},
	}

	buf := make([]byte, 1024)
	if _, err := rand.New(rand.NewSource(time.Now().UnixNano())).Read(buf); err != nil {
		t.Fatalf("unexpected error while randomizing input: %v\n", err)
	}
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/feed", bytes.NewReader(buf))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	svc.addFeed(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status code %d, got %d\n", http.StatusBadRequest, rr.Code)
	}
}

func TestFeedHandlerInvalidFieldValue(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	svc := Service{
		log:   log.New(f, "", log.LstdFlags),
		feeds: map[string]string{},
	}

	buf, err := json.Marshal(struct {
		Name string  `json:"name"`
		URL  float64 `json:"url"`
	}{
		Name: "test-feed",
		URL:  42.42,
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/feed", bytes.NewReader(buf))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	svc.addFeed(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status code %d, got %d\n", http.StatusBadRequest, rr.Code)
	}
}

func TestFeedHandlerExtraField(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	svc := Service{
		log:   log.New(f, "", log.LstdFlags),
		feeds: map[string]string{},
	}

	buf, err := json.Marshal(struct {
		Feed
		Invalid string `json:"invalid"`
	}{
		Feed: Feed{
			Name: "test-feed",
			URL:  "test-feed-url",
		},
		Invalid: "invalid-field",
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/feed", bytes.NewReader(buf))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	svc.addFeed(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status code %d, got %d\n", http.StatusBadRequest, rr.Code)
	}
}

func TestFeedHandlerEmptyBody(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	svc := Service{
		log:   log.New(f, "", log.LstdFlags),
		feeds: map[string]string{},
	}

	req, err := http.NewRequest(http.MethodPost, "/feed", bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	svc.addFeed(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status code %d, got %d\n", http.StatusBadRequest, rr.Code)
	}
}

func TestFeedHandlerBodyTooLarge(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	svc := Service{
		log:   log.New(f, "", log.LstdFlags),
		feeds: map[string]string{},
	}

	feeds := make([]Feed, 1024*1024)
	for i := 0; i < 1024*1024; i++ {
		feeds[i].Name = fmt.Sprintf("test-too-large-%d", i)
		feeds[i].URL = fmt.Sprintf("test-too-large-url-%d", i)
	}
	buf, err := json.Marshal(feeds)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/feed", bytes.NewReader(buf))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	svc.addFeed(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status code %d, got %d\n", http.StatusRequestEntityTooLarge, rr.Code)
	}
}

func TestFeedHandlerMultipleObjects(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	svc := Service{
		log:   log.New(f, "", log.LstdFlags),
		feeds: map[string]string{},
	}

	buf, err := json.Marshal([]Feed{
		{
			Name: "test-1",
			URL:  "test-url-1",
		},
		{
			Name: "test-2",
			URL:  "test-url-2",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/feed", bytes.NewReader(buf))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	svc.addFeed(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status code %d, got %d\n", http.StatusBadRequest, rr.Code)
	}
}

func TestFeedHandler(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	svc := Service{
		log:   log.New(f, "", log.LstdFlags),
		feeds: map[string]string{},
	}

	buf, err := json.Marshal(Feed{
		Name: "test-1",
		URL:  "test-url-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/feed", bytes.NewReader(buf))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	svc.addFeed(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status code %d, got %d\n", http.StatusAccepted, rr.Code)
	}

	if svc.feeds["test-1"] != "test-url-1" {
		t.Fatalf("new feed (name %s, url %s) expected, got url %s\n", "test-1", "test-url-1", svc.feeds["test-1"])
	}
}

func TestFeedHandlerConcurrentRequests(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	svc := Service{
		log:   log.New(f, "", log.LstdFlags),
		feeds: map[string]string{},
	}

	var wgStart, wgEnd sync.WaitGroup

	wgStart.Add(10)
	wgEnd.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer wgEnd.Done()

			buf, err := json.Marshal(Feed{
				Name: "test-concurrent",
				URL:  fmt.Sprintf("test-concurrent-url-%d", i),
			})
			if err != nil {
				t.Error(err)
			}
			req, err := http.NewRequest(http.MethodPost, "/feed", bytes.NewReader(buf))
			if err != nil {
				t.Error(err)
			}
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// wait for all goroutines to issue the requests concurrently
			wgStart.Done()
			wgStart.Wait()

			svc.addFeed(rr, req)

			if rr.Code != http.StatusAccepted {
				t.Errorf("expected status code %d, got %d\n", http.StatusAccepted, rr.Code)
			}
		}(i)
	}

	// wait for all goroutines to finish
	wgEnd.Wait()
}

type MockedAddr struct{}

func (ma MockedAddr) Network() string { return "" }
func (ma MockedAddr) String() string  { return "" }

type MockedConnection struct{}

func (mc MockedConnection) Read(b []byte) (n int, err error)   { return 0, nil }
func (mc MockedConnection) Write(b []byte) (n int, err error)  { return 0, nil }
func (mc MockedConnection) Close() error                       { return nil }
func (mc MockedConnection) LocalAddr() net.Addr                { return MockedAddr{} }
func (mc MockedConnection) RemoteAddr() net.Addr               { return MockedAddr{} }
func (mc MockedConnection) SetDeadline(t time.Time) error      { return nil }
func (mc MockedConnection) SetReadDeadline(t time.Time) error  { return nil }
func (mc MockedConnection) SetWriteDeadline(t time.Time) error { return nil }

func TestItemsHandler(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	svc := Service{
		log:   log.New(f, "", log.LstdFlags),
		feeds: map[string]string{},
	}

	ctx := context.WithValue(
		context.Background(),
		DeadlineControllerKey,
		NewDeadlineController(MockedConnection{}, svc.server),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/items", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add our context to the request: note that WithContext returns a copy of
	// the request, which we must assign.
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	svc.streamItems(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d\n", http.StatusAccepted, rr.Code)
	}

	if rr.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatalf("expected Content-Type %s, got %s\n", "text/event-stream", rr.Header().Get("Content-Type"))
	}
	if rr.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("expected Cache-Control %s, got %s\n", "no-cache", rr.Header().Get("Cache-Control"))
	}
	if rr.Header().Get("Connection") != "keep-alive" {
		t.Fatalf("expected Connection %s, got %s\n", "keep-alive", rr.Header().Get("Connection"))
	}
	if rr.Header().Get("Transfer-Encoding") != "chunked" {
		t.Fatalf("expected Transfer-Encoding %s, got %s\n", "chunked", rr.Header().Get("Transfer-Encoding"))
	}
}
