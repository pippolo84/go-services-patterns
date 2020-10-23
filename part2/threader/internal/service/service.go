package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// Options holds all the configuration options for the Service
type Options struct {
	Name string

	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	Handler http.Handler
	Logger  *log.Logger
}

// Service is the type of a rss service
type Service struct {
	name   string
	server *http.Server

	Logger *log.Logger
}

// NewService builds a new Service using the config
// options passed as a parameter
func NewService(opts Options) *Service {
	return &Service{
		server: &http.Server{
			Addr:         opts.Addr,
			Handler:      opts.Handler,
			ReadTimeout:  opts.ReadTimeout,
			WriteTimeout: opts.WriteTimeout,
			IdleTimeout:  opts.IdleTimeout,
		},
		Logger: opts.Logger,
	}
}

const (
	// SrvReadTimeout is the http server read timeout
	SrvReadTimeout time.Duration = 5 * time.Second
	// SrvWriteTimeout is the http server write timeout
	SrvWriteTimeout time.Duration = 10 * time.Second
	// SrvIdleTimeout is the http server idle timeout
	SrvIdleTimeout time.Duration = 60 * time.Second
)

// NewDefaultService builds a new Service using (opinionated) default configuration options
// The mandatyory parameters are the name of the service, the address on which it will expose itself
// and the http handler
func NewDefaultService(name, addr string, handler http.Handler) *Service {
	return &Service{
		name: name,
		server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  SrvReadTimeout,
			WriteTimeout: SrvWriteTimeout,
			IdleTimeout:  SrvIdleTimeout,
		},
		Logger: log.New(os.Stdout, fmt.Sprintf("%s: ", name), log.LstdFlags),
	}
}

// Run starts the service
// is about to call ListenAndServer
// It returns a channel where all the errors are forwarded
func (svc *Service) Run() <-chan error {
	svcErrs := make(chan error)

	go func() {
		svc.Logger.Printf("running on %s", svc.server.Addr)

		defer func() {
			close(svcErrs)
			svc.Logger.Println("run stopped")
		}()

		if err := svc.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			svc.Logger.Printf("run error: %v\n", err)
			svcErrs <- err
		}
	}()

	return svcErrs
}

// Shutdown shuts down the service
// It takes a context that it's passed to the underling HTTP server to control the shutdown process
// and a *sync.WaitGroup to signal when the shutdown process is ended
func (svc *Service) Shutdown(ctx context.Context, wg *sync.WaitGroup) error {
	svc.Logger.Println("shutdown")
	defer func() {
		wg.Done()
		svc.Logger.Println("bye")
	}()

	return svc.server.Shutdown(ctx)
}

// Feed holds information about a RSS feed
type Feed struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// News holds information about a single item from a RSS feed
type News struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
}
