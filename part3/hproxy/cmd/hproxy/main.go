package main

import (
	"context"
	"log"
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
	Name string = "hrpoxy"

	// Address is the string representation of the address where the service will listen
	Address string = ":8081"

	// SvcCooldownTimeout is the maximum cooldown time before forcing the shutdown
	SvcCooldownTimeout time.Duration = 10 * time.Second
)

func main() {
	proxy := NewProxy()

	if err := proxy.Init(); err != nil {
		log.Fatalf("initialization error: %v\n", err)
	}
	defer proxy.Close()

	errs := proxy.Run()

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

	if err := proxy.Shutdown(shutdownCtx, &stopWg); err != nil {
		log.Println(err)
	}

	// wait for service to cleanup everything
	stopWg.Wait()
}

// Proxy represents our specific service
type Proxy struct {
	*service.Service
}

// NewProxy returns a new Proxy ready to be initialized and run
func NewProxy() *Proxy {
	proxy := &Proxy{}

	router := mux.NewRouter()

	router.HandleFunc("/", proxy.handler).Schemes("http").Methods(http.MethodGet)

	proxy.Service = service.NewDefaultService(Name, Address, router)

	return proxy
}

// Init initializes private components of the service
func (p *Proxy) Init() error {
	return nil
}

// Close frees all resources used by the service private components
func (p *Proxy) Close() {
}

// ************************** HTTP Handlers **************************

func (p *Proxy) handler(w http.ResponseWriter, r *http.Request) {

}
