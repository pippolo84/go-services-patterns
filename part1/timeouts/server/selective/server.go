package selective

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// Timeouts holds all the timeouts to be set in the Server
type Timeouts struct {
	Write time.Duration
}

// Server is a HTTP server with configurable timeouts
// and some defaults handlers to experiment with them
type Server struct {
	http.Server
}

// DeadlineController packs the underlying stream-oriented connection
// with a reference to the server serving the HTTP request
// see https://github.com/golang/go/issues/16100
type DeadlineController struct {
	c net.Conn
	s *http.Server
}

// ContextKey is a custom type for our deadline controller key value
type ContextKey string

// DeadlineControllerKey is the name of the value passed through context
const DeadlineControllerKey ContextKey = "deadline-controller"

// NewDeadlineController returns a type that holds the underlying TCP connection
// and a reference to the server
func NewDeadlineController(c net.Conn, s *http.Server) *DeadlineController {
	return &DeadlineController{
		c: c,
		s: s,
	}
}

// ExtendWriteDeadline extends the write deadline on the underlying
// stream-oriented connection
func (dc *DeadlineController) ExtendWriteDeadline() error {
	return dc.c.SetWriteDeadline(time.Now().Add(dc.s.WriteTimeout))
}

// NewServer returns a HTTP server listening on addr
// and configured with the specified timeouts
func NewServer(addr string, timeouts Timeouts) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler)
	mux.HandleFunc("/timeout", timeoutHandler)

	srv := &Server{
		http.Server{
			Addr:         addr,
			WriteTimeout: timeouts.Write,
			Handler:      mux,
		},
	}

	srv.ConnContext = func(ctx context.Context, c net.Conn) context.Context {
		dc := NewDeadlineController(c, &srv.Server)
		return context.WithValue(ctx, DeadlineControllerKey, dc)
	}

	return srv
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
	count := 0
	for {
		time.Sleep(time.Second)

		value := r.Context().Value(DeadlineControllerKey)
		dlController, ok := value.(*DeadlineController)
		if !ok {
			log.Println("casting errror")
			return
		}

		// extend the write timeout of the underlying TCP connection
		if err := dlController.ExtendWriteDeadline(); err != nil {
			log.Println(err)
			return
		}

		count++

		if count > 10 {
			break
		}
	}

	fmt.Fprintf(w, "Hello, world!")
}

func timeoutHandler(w http.ResponseWriter, r *http.Request) {
	// this will timeout due to WriteTimeout
	time.Sleep(5 * time.Second)

	fmt.Fprintf(w, "Hello, world!")
}
