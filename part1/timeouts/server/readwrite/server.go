package readwrite

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Server Timeout: https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/

// Timeouts holds all the timeouts to be set in the Server
type Timeouts struct {
	Read  time.Duration
	Write time.Duration

	Handler time.Duration
}

// Server is a HTTP server with configurable timeouts
// and some defaults handlers to experiment with them
type Server struct {
	http.Server
}

// NewServer returns a HTTP server listening on addr
// and configured with the specified timeouts
func NewServer(addr string, timeouts Timeouts) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler)
	mux.Handle("/slow", http.TimeoutHandler(http.HandlerFunc(slowHandler), timeouts.Handler, "timeout"))
	mux.HandleFunc("/timeout", timeoutHandler)

	return &Server{
		http.Server{
			Addr: addr,

			ReadTimeout:  timeouts.Read,
			WriteTimeout: timeouts.Write,

			Handler: mux,
		},
	}
}

// Run starts the server, making it listening on specified address
// it returns a channel where all errors are relayed
func (srv *Server) Run() <-chan error {
	errs := make(chan error)

	go func() {
		defer close(errs)
		if err := srv.ListenAndServe(); err != nil {
			errs <- err
		}
	}()

	return errs
}

// Shutdown makes the server stop listening and refuse further connections
// It takes a context to limit the shutdown duration and a wait group to signal
// the caller when the shutdown process has finished
func (srv *Server) Shutdown(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	return srv.Server.Shutdown(ctx)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	fmt.Fprintf(w, "Hello, slow world!")
}

func timeoutHandler(w http.ResponseWriter, r *http.Request) {
	// Despite returning a io.EOF on client side, the handler won't stop execution.
	// Check the duration of TestWriteTimeout test to prove it!
	time.Sleep(6 * time.Second)
	fmt.Fprintf(w, "Hello, timeout world!")
}
