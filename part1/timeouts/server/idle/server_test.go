package idle

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
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

func TestIdleTimeout(t *testing.T) {
	srv, srvErrs := startServer(t, ":8080", Timeouts{
		Idle: 2 * time.Second,
	})

	getRequest := func() string {
		t.Helper()

		res, err := http.Get(fmt.Sprintf("http://localhost%s", srv.Addr))
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		slurp, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}

		return string(slurp)
	}

	addr1, addr2 := getRequest(), getRequest()
	if addr1 != addr2 {
		t.Fatalf("requests were on different connections")
	}

	// wait a little longer than idle timeout
	time.Sleep(3 * time.Second)

	addr3 := getRequest()
	if addr2 == addr3 {
		t.Fatal("request 3 unexpectedly on the same TCP connection")
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
