package readwrite

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// go test -v -timeout=20s .

const (
	SrvReadTimeout  time.Duration = 5 * time.Second
	SrvWriteTimeout time.Duration = 5 * time.Second

	HandlerTimeout time.Duration = time.Second
)

func TestMain(m *testing.M) {
	// create and run the server under test
	srv := NewServer(":8080", Timeouts{
		Read:  SrvReadTimeout,
		Write: SrvWriteTimeout,

		Handler: HandlerTimeout,
	})

	go func() {
		for err := range srv.Run() {
			log.Fatal(err)
		}
	}()

	// wait for the server to listen on port 8080
	for {
		_, err := net.Dial("tcp", "127.0.0.1:8080")
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "connection refused") {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestReadTimeout(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if _, err := conn.Write(nil); err != nil {
		t.Fatal(err)
	}

	// server should close the connection before client ReadDeadline
	if err := conn.SetReadDeadline(time.Now().Add(2 * SrvReadTimeout)); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 512)
	if _, err := conn.Read(buf); !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
}

func TestWriteTimeout(t *testing.T) {
	_, err := http.Get("http://localhost:8080/timeout")
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF error, got: %v\n", err)
	}
}

func TestHandlerTimeout(t *testing.T) {
	res, err := http.Get("http://localhost:8080/slow")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status code %d, got %d\n", http.StatusServiceUnavailable, res.StatusCode)
	}
}

func TestGetRequest(t *testing.T) {
	res, err := http.Get("http://localhost:8080/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d\n", http.StatusOK, res.StatusCode)
	}
}
