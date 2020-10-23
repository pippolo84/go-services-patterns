package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
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
	Name string = "broker"

	// Address is the string representation of the address where the service will listen
	Address string = ":8080"

	// SvcCooldownTimeout is the maximum cooldown time before forcing the shutdown
	SvcCooldownTimeout time.Duration = 10 * time.Second

	// MaxRequestBodySize is the maximum size, in bytes, if a request body
	MaxRequestBodySize int64 = 1 << 20
)

func main() {
	broker := NewBroker()

	if err := broker.Init(); err != nil {
		log.Fatalf("initialization error: %v\n", err)
	}
	defer broker.Close()

	errs := broker.Run()

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

	if err := broker.Shutdown(shutdownCtx, &stopWg); err != nil {
		log.Println(err)
	}

	// wait for service to cleanup everything
	stopWg.Wait()
}

// Broker represents our specific service
type Broker struct {
	*service.Service
	store *model.Objects
	mu    sync.Mutex
}

// NewBroker returns a new Broker ready to be initialized and run
func NewBroker() *Broker {
	broker := &Broker{}

	router := mux.NewRouter()

	router.HandleFunc("/thread", broker.threads).Schemes("http").Methods(http.MethodGet)
	router.HandleFunc("/thread/{tid}", broker.getThread).Schemes("http").Methods(http.MethodGet)
	router.HandleFunc("/thread/{tid}", broker.newThread).Schemes("http").Methods(http.MethodPost)

	router.HandleFunc("/message/{tid}/{mid}", broker.newMessage).Schemes("http").Methods(http.MethodPost)

	router.HandleFunc("/upvote/{tid}/{mid}", broker.upvote).Schemes("http").Methods(http.MethodPatch)

	broker.Service = service.NewDefaultService(Name, Address, router)
	broker.store = model.NewObjectsStore()

	return broker
}

// Init initializes private components of the service, like its internal broker
func (b *Broker) Init() error {
	if err := b.store.Init(); err != nil {
		return err
	}

	return nil
}

// Close frees all resources used by the service private components
func (b *Broker) Close() {
	b.store.Close()
}

// ************************** HTTP Handlers **************************

func (b *Broker) threads(w http.ResponseWriter, r *http.Request) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ths := b.store.Threads()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ths); err != nil {
		_ = json.NewEncoder(w).Encode(struct {
			Code int    `json:"code"`
			Text string `json:"text"`
		}{
			Code: http.StatusInternalServerError,
			Text: http.StatusText(http.StatusInternalServerError),
		})
	}
}

func (b *Broker) getThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	t, err := b.store.GetThread(vars["tid"])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(t); err != nil {
		_ = json.NewEncoder(w).Encode(struct {
			Code int    `json:"code"`
			Text string `json:"text"`
		}{
			Code: http.StatusInternalServerError,
			Text: http.StatusText(http.StatusInternalServerError),
		})
	}
}

func (b *Broker) newThread(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxRequestBodySize))
	dec.DisallowUnknownFields()

	var t model.Thread
	if err := validate.JSON(r.Body, &t); err != nil {
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

	// thread topic is mandatory
	if t.Topic == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)

	t.ID = vars["tid"]
	t.Messages = nil

	b.mu.Lock()
	defer b.mu.Unlock()

	b.store.AddThread(&t)

	w.WriteHeader(http.StatusCreated)
}

func (b *Broker) newMessage(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxRequestBodySize))
	dec.DisallowUnknownFields()

	var m model.Message
	if err := validate.JSON(r.Body, &m); err != nil {
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

	// text is mandatory
	if m.Text == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)

	m.ID = vars["mid"]
	m.ThreadID = vars["tid"]
	m.Timestamp = time.Now().Format(time.RFC3339)
	m.Votes = 0

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	resp, err := b.updateScore(ctx, m.ThreadID, m.ID, m.Votes)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			http.Error(w, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.store.AddMessage(m.ThreadID, &m); err != nil {
		var e model.ErrThreadIDNotFound
		if errors.As(err, &e) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (b *Broker) upvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	tid := vars["tid"]
	mid := vars["mid"]

	b.mu.Lock()
	defer b.mu.Unlock()

	m, err := b.store.GetMessage(tid, mid)
	if err != nil {
		var (
			e1 model.ErrThreadIDNotFound
			e2 model.ErrMessageIDNotFound
		)
		if errors.As(err, &e1) || errors.As(err, &e2) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	m.Votes++

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	resp, err := b.updateScore(ctx, tid, mid, m.Votes)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			http.Error(w, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}
}

func (b *Broker) updateScore(ctx context.Context, tid, mid string, score int) (*http.Response, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(model.Score{
		ThreadID:  tid,
		MessageID: mid,
		Votes:     score,
	}); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:8081/score", bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
