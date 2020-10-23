package graceful

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

// Server is a HTTP server that supports graceful shutdown
type Server struct {
	http.Server
}

// NewServer returns a reference to a new Server listening on addr
func NewServer(addr string) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler)
	mux.HandleFunc("/slow", slowHandler)
	// see https://github.com/hashrocket/ws for a websocket client
	mux.HandleFunc("/ws", wsHandler)

	ctx, cancel := context.WithCancel(context.Background())
	srv := &Server{
		Server: http.Server{
			Addr:    addr,
			Handler: mux,
			// use the context with cancellation to be able to gracefully shutdown websocket connections
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
		},
	}

	// on shutdown, cancel the context
	srv.Server.RegisterOnShutdown(cancel)

	return srv
}

// Shutdown makes the server stop listening and refuse further connections
// It takes a context to limit the shutdown duration and a wait group to signal
// the caller when the shutdown process has finished
func (srv *Server) Shutdown(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	return srv.Server.Shutdown(ctx)
}

// Run starts the server, making it listening on specified address
// it returns a channel where all errors are relayed
func (srv *Server) Run() <-chan error {
	errs := make(chan error)

	go func() {
		defer close(errs)

		log.Printf("server: start listening on %s\n", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}

		log.Println("server: bye!")
	}()

	return errs
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	// Shutdown won't close this connection until it returns to idle
	time.Sleep(10 * time.Second)
	fmt.Fprintf(w, "Hello, slow world!")
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		fmt.Printf("websocket upgrade error: %v\n", err)
	}
	// Additional calls to Close are no-ops
	defer c.Close(websocket.StatusInternalError, "internal server error")

	// to cancel the write in case of a slow reader
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()

	if err := c.Write(ctx, websocket.MessageText, []byte("Hello, ws world!")); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("ws write error: %v\n", err)
		}
		return
	}

	mtype, buf, err := c.Read(ctx)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("ws read error: %v\n", err)
		}
		return
	}

	if mtype == websocket.MessageText {
		fmt.Printf("received: %s\n", string(buf))
	}

	c.Close(websocket.StatusNormalClosure, "")
}
