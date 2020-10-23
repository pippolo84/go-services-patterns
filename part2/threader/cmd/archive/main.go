package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Pippolo84/go-services-patterns/part2/threader/internal/model"
	"github.com/Pippolo84/go-services-patterns/part2/threader/internal/service"
	"github.com/Pippolo84/go-services-patterns/part2/threader/internal/validate"
	"github.com/gorilla/mux"
)

const (
	// Name is the name of the service
	Name string = "archive"

	// Address is the string representation of the address where the service will listen
	Address string = ":8081"

	// SvcCooldownTimeout is the maximum cooldown time before forcing the shutdown
	SvcCooldownTimeout time.Duration = 10 * time.Second

	// MaxRequestBodySize is the maximum size, in bytes, if a request body
	MaxRequestBodySize int64 = 1 << 20
)

func main() {
	archive := NewArchive()

	if err := archive.Init(); err != nil {
		log.Fatalf("initialization error: %v\n", err)
	}
	defer archive.Close()

	errs := archive.Run()

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

	if err := archive.Shutdown(shutdownCtx, &stopWg); err != nil {
		log.Println(err)
	}

	// wait for service to cleanup everything
	stopWg.Wait()
}

// Archive represents our specific service
type Archive struct {
	*service.Service
	store   *model.Analytics
	nfail   int
	restore time.Time
	mu      sync.Mutex
}

// NewArchive returns a new Archive ready to be initialized and run
func NewArchive() *Archive {
	archive := &Archive{}

	router := mux.NewRouter()

	router.HandleFunc("/score", archive.newScore).Schemes("http").Methods(http.MethodPost)
	router.HandleFunc("/highest-score", archive.highestScore).Schemes("http").Methods(http.MethodGet)

	archive.Service = service.NewDefaultService(Name, Address, router)
	archive.store = model.NewAnalyticsStore()

	return archive
}

// Init initializes private components of the service, like its internal store
func (a *Archive) Init() error {
	rand.Seed(time.Now().UnixNano())

	if err := a.store.Init(); err != nil {
		return err
	}

	return nil
}

// Close frees all resources used by the service private components
func (a *Archive) Close() {
	a.store.Close()
}

// ************************** HTTP Handlers **************************

func (a *Archive) newScore(w http.ResponseWriter, r *http.Request) {
	if err := a.updateStatus(); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// simulate a failure
	if rand.Float64() >= 0.90 {
		a.fail()

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxRequestBodySize))
	dec.DisallowUnknownFields()

	var s model.Score
	if err := validate.JSON(r.Body, &s); err != nil {
		switch err.(type) {
		case validate.ErrJSONSyntax:
		case validate.ErrInvalidValue:
		case validate.ErrUnknownField:
		case validate.ErrEmpty:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case validate.ErrSzExceeded:
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		case validate.ErrGeneric:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.store.AddScore(&s)
}

func (a *Archive) highestScore(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	score := a.store.HighestScore()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(score); err != nil {
		_ = json.NewEncoder(w).Encode(struct {
			Code int    `json:"code"`
			Text string `json:"text"`
		}{
			Code: http.StatusInternalServerError,
			Text: http.StatusText(http.StatusInternalServerError),
		})
	}
}

func (a *Archive) fail() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// simulate a transient failure
	extend := 1 + rand.Intn(10)
	a.restore = time.Now().Add(time.Duration(extend) * time.Millisecond)
	a.nfail = 1

}

func (a *Archive) updateStatus() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if time.Now().After(a.restore) {
		// restore service
		a.nfail = 0
		return nil
	}

	// simulate a progressive worsening of the service health status
	a.nfail = (a.nfail + 1) % 100
	extend := a.nfail + 1 + rand.Intn(10)
	a.restore = a.restore.Add(time.Duration(extend) * time.Millisecond)

	// keep the client stuck if consecutive failures are high
	if rand.Intn(a.nfail) >= 90 {
		interval := time.Duration(rand.Intn(100)) * time.Millisecond
		if time.Now().Add(interval).After(a.restore) {
			interval = time.Until(a.restore)
		}
		time.Sleep(interval)
	}

	return errors.New("service failure")
}
