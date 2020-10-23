package streaming

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// go test -v -timeout=20s .

func startServer(t *testing.T, addr string, timeouts Timeouts) (*Server, <-chan error) {
	t.Helper()

	srv := NewServer(addr, timeouts)

	errs := make(chan error)

	go func() {
		defer close(errs)

		for err := range srv.Run() {
			errs <- err
		}
	}()

	// wait for the server to listen on addr
	for {
		_, err := net.Dial("tcp", fmt.Sprintf("localhost%s", addr))
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "connection refused") {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		log.Fatal(err)
	}

	return srv, errs
}

func TestWriteTimeout(t *testing.T) {
	srv, srvErrs := startServer(t, ":8080", Timeouts{
		Write: 2 * time.Second,
	})

	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost%s/streaming", srv.Addr), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected http status code %d, got %d\n", http.StatusOK, res.StatusCode)
	}

	buf := make([]byte, 512)
	for {
		_, err := res.Body.Read(buf)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Fatal(err)
		}

		if errors.Is(err, io.ErrUnexpectedEOF) {
			break
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	if err := srv.Shutdown(context.Background(), &wg); err != nil {
		t.Fatal(err)
	}

	wg.Wait()

	// srvErrs should have been closed without errors
	for err := range srvErrs {
		if errors.Is(err, http.ErrServerClosed) {
			continue
		}
		t.Fatal(err)
	}
}

func TestEOF(t *testing.T) {
	srv, srvErrs := startServer(t, ":8080", Timeouts{})

	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost%s/streaming", srv.Addr), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected http status code %d, got %d\n", http.StatusOK, res.StatusCode)
	}

	buf := make([]byte, 512)
	for {
		_, err := res.Body.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			t.Fatal(err)
		}

		// connection won't timeout and we'll reach the End Of Stream
		if errors.Is(err, io.EOF) {
			break
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	if err := srv.Shutdown(context.Background(), &wg); err != nil {
		t.Fatal(err)
	}

	wg.Wait()

	// srvErrs should have been closed without errors
	for err := range srvErrs {
		if errors.Is(err, http.ErrServerClosed) {
			continue
		}

		t.Fatal(err)
	}
}
