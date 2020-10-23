package selective

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

const (
	SrvWriteTimeout time.Duration = 2 * time.Second
)

func TestMain(m *testing.M) {
	// create and run the server under test
	srv := NewServer(":8080", Timeouts{
		Write: SrvWriteTimeout,
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

func TestWriteTimeout(t *testing.T) {
	_, err := http.Get("http://localhost:8080/timeout")
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF error, got: %v\n", err)
	}
}

func TestWriteTimeoutExtended(t *testing.T) {
	_, err := http.Get("http://localhost:8080/")
	if err != nil {
		t.Fatalf("expected no error, got: %v\n", err)
	}
}
