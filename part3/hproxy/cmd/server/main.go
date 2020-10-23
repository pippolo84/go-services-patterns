package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Pippolo84/go-services-patterns/part3/hproxy/internal/service"
	"github.com/gorilla/mux"
)

const (
	// Name is the name of the service
	Name string = "server"

	// Address is the string representation of the address where the service will listen
	Address string = ":8080"

	// SvcCooldownTimeout is the maximum cooldown time before forcing the shutdown
	SvcCooldownTimeout time.Duration = 10 * time.Second
)

func main() {
	server := NewServer()

	if err := server.Init(); err != nil {
		log.Fatalf("initialization error: %v\n", err)
	}
	defer server.Close()

	errs := server.Run()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// block until a signal or an error is received
	select {
	case err := <-errs:
		log.Println(err)
	case sig := <-signalChan:
		log.Printf("got signal: %v, shutting down...\n", sig)
	}

	// graceful shutdown the service
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), SvcCooldownTimeout)
	defer cancelShutdown()

	var stopWg sync.WaitGroup
	stopWg.Add(1)

	if err := server.Shutdown(shutdownCtx, &stopWg); err != nil {
		log.Println(err)
	}

	// wait for service to cleanup everything
	stopWg.Wait()
}

// Server represents our specific service
type Server struct {
	*service.Service
}

// NewServer returns a new Server ready to be initialized and run
func NewServer() *Server {
	server := &Server{}

	router := mux.NewRouter()

	router.HandleFunc("/", server.handler).Schemes("http").Methods(http.MethodGet)

	server.Service = service.NewDefaultService(Name, Address, router)

	return server
}

// Init initializes private components of the service, like its internal server
func (s *Server) Init() error {
	rand.Seed(time.Now().UnixNano())
	return nil
}

// Close frees all resources used by the service private components
func (s *Server) Close() {
}

// ************************** HTTP Handlers **************************

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	x := rand.Float32()
	switch {
	case x >= 0.999:
		time.Sleep(1000 * time.Millisecond)
	case x >= 0.99:
		time.Sleep(800 * time.Millisecond)
	case x >= 0.95:
		time.Sleep(80 * time.Millisecond)
	case x >= 0.5:
		time.Sleep(10 * time.Millisecond)
	default:
		time.Sleep(5 * time.Millisecond)
	}

	fmt.Fprintln(w, "Hi, I'm the high latency server!")
}
