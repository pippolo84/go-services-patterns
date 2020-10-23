package graceful

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// FIXME: add a test for the ws handler

// go test -v -timeout=20s .

func startServer(t *testing.T, addr string) (*Server, <-chan error) {
	t.Helper()

	srv := NewServer(addr)

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

func TestGetRequest(t *testing.T) {
	defer goleak.VerifyNone(t)

	srv, srvErrs := startServer(t, ":8080")

	res, err := http.Get("http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d\n", http.StatusOK, res.StatusCode)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	if err := srv.Shutdown(context.Background(), &wg); err != nil {
		t.Fatal(err)
	}

	wg.Wait()

	// srvErrs should have been closed without errors
	for err := range srvErrs {
		t.Fatal(err)
	}
}

func TestGracefulShutdown(t *testing.T) {
	defer goleak.VerifyNone(t)

	srv, srvErrs := startServer(t, ":8080")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelShutdown()

	var wg sync.WaitGroup
	wg.Add(1)

	if err := srv.Shutdown(shutdownCtx, &wg); err != nil {
		t.Fatal(err)
	}

	wg.Wait()

	// srvErrs should have been closed without errors
	for err := range srvErrs {
		t.Fatal(err)
	}
}

func slowRequest(t *testing.T, addr string) <-chan error {
	t.Helper()

	reqErrs := make(chan error)

	go func() {
		defer close(reqErrs)

		var err *net.OpError
		if _, e := http.Get(addr); !errors.As(e, &err) {
			reqErrs <- err
		}
	}()

	return reqErrs
}

func TestGracefulShutdownTimeout(t *testing.T) {
	srv, srvErrs := startServer(t, ":8081")

	reqErrs := slowRequest(t, "http://localhost:8081/slow")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), time.Second)
	defer cancelShutdown()

	var wg sync.WaitGroup
	wg.Add(1)

	if err := srv.Shutdown(shutdownCtx, &wg); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got: %v\n", err)
	}

	wg.Wait()

	// srvErrs and reqErrs should have been closed without errors
	for err := range srvErrs {
		t.Fatal(err)
	}
	for err := range reqErrs {
		t.Fatal(err)
	}
}
