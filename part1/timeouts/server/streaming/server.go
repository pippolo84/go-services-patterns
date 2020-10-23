package streaming

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Timeouts holds all the timeouts to be set in the Server
type Timeouts struct {
	Write time.Duration
}

// Server is a HTTP server with configurable write timeout
// and some defaults handlers to experiment with them
type Server struct {
	http.Server
}

// NewServer returns a HTTP server listening on addr
// and configured with the specified timeouts
func NewServer(addr string, timeouts Timeouts) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/streaming", streamingHandler)

	return &Server{
		http.Server{
			Addr: addr,

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

func streamingHandler(w http.ResponseWriter, r *http.Request) {
	defer log.Println("streaming finished!")

	// make sure the connection supports the streaming
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// set the headers or event streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	count := 0
	for range ticker.C {
		fmt.Fprintln(w, "Hello, streaming world!")
		f.Flush()

		count++
		if count > 5 {
			break
		}
	}
}
