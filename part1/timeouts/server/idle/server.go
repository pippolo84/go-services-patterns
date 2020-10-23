package idle

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

// Timeouts holds all the timeouts to be set in the Server
type Timeouts struct {
	Idle time.Duration
}

// Server is a HTTP server with configurable idle timeout
// and some defaults handlers to experiment with them
type Server struct {
	http.Server
}

// NewServer returns a HTTP server listening on addr
// and configured with the specified timeouts
func NewServer(addr string, timeouts Timeouts) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler)

	return &Server{
		http.Server{
			Addr: addr,

			IdleTimeout: timeouts.Idle,

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
	if _, err := io.Copy(ioutil.Discard, r.Body); err != nil {
		log.Println(err)
	}
	if _, err := io.WriteString(w, r.RemoteAddr); err != nil {
		log.Println(err)
	}
}
