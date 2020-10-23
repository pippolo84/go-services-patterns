// +build integration

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// test with `go test --tags=integration -v -race -timeout=120s .`

func startServer(t *testing.T, addr string) (*Service, <-chan error) {
	t.Helper()

	svc := NewService(addr)

	errs := make(chan error)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer close(errs)

		for err := range svc.Run(&wg) {
			errs <- err
		}
	}()

	// wait service init phase
	wg.Wait()

	// wait for the service to listen on addr
	for {
		_, err := net.Dial("tcp", addr)
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "connection refused") {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		log.Fatal(err)
	}

	return svc, errs
}

func TestGracefulShutdown(t *testing.T) {
	defer goleak.VerifyNone(t)

	svc, svcErrs := startServer(t, "localhost:8080")

	var wg sync.WaitGroup
	wg.Add(1)

	if err := svc.Shutdown(context.Background(), &wg); err != nil {
		t.Fatal(err)
	}

	wg.Wait()

	// svcErrs should have been closed without errors
	for err := range svcErrs {
		t.Fatal(err)
	}
}

func TestExtendWriteDeadline(t *testing.T) {
	defer goleak.VerifyNone(t)

	svc, svcErrs := startServer(t, "localhost:8080")

	// add a feed
	buf, err := json.Marshal(Feed{
		Name: "test-feed",
		URL:  "http://joeroganexp.joerogan.libsynpro.com/rss",
	})
	if err != nil {
		t.Fatal(err)
	}
	feedResp, err := http.Post("http://localhost:8080/feed", "application/json", bytes.NewReader(buf))
	if err != nil {
		t.Fatal(err)
	}
	defer feedResp.Body.Close()

	if feedResp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status code %d, got %d\n", http.StatusAccepted, feedResp.StatusCode)
	}

	// request feed streaming
	req, err := http.NewRequest("GET", "http://localhost:8080/items", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")

	itemsResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer itemsResp.Body.Close()

	if itemsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected http status code %d, got %d\n", http.StatusOK, itemsResp.StatusCode)
	}

	// check that the stream lasts longer than WriteTimeout
	stop := time.After(15 * time.Second)
done:
	for {
		select {
		case <-stop:
			break done
		default:
			// TODO: mock the RSS service to use a well-known bytes number in io.CopyN
			// and to avoid issues if RSS server is down
			// or try this from DockerHub (use testcontainers-go)
			_, err := io.CopyN(ioutil.Discard, itemsResp.Body, 1024*1024)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	// goleak doesn't play well with t.Cleanup, so we add the cleanup
	// part manually at the end of the test
	var wg sync.WaitGroup
	wg.Add(1)

	if err := svc.Shutdown(context.Background(), &wg); err != nil {
		t.Fatal(err)
	}

	wg.Wait()

	// svcErrs should have been closed without errors
	for err := range svcErrs {
		t.Fatal(err)
	}
}
