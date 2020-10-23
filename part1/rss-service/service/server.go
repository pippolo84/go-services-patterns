package service

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// Service is the type of a rss service
type Service struct {
	server *http.Server
	router *mux.Router
}

// NewService builds a new Service ready to be run
func NewService(addr string) *Service {
	svc := &Service{
		server: &http.Server{
			Addr: addr,
		},
		router: mux.NewRouter(),
	}
	svc.server.Handler = svc.router
	svc.router.HandleFunc("/feeds", svc.getFeeds).Schemes("http").Methods(http.MethodGet)
	svc.router.HandleFunc("/feed", svc.addFeed).Schemes("http").Methods(http.MethodPost)
	svc.router.HandleFunc("/items", svc.streamItems).Schemes("http").Methods(http.MethodGet)

	return svc
}

// Run starts the service
// It returns a channel where all the errors are forwarded
func (svc *Service) Run(wg *sync.WaitGroup) <-chan error {
	errs := make(chan error)

	return errs
}

// Shutdown shuts down the service
// It takes a context that it's passed to the underling HTTP server to control the shutdown process
// and a *sync.WaitGroup to signal when the shutdown process is ended
func (svc *Service) Shutdown(ctx context.Context, wg *sync.WaitGroup) error {
}

// DeadlineController holds a reference to the underlying TCP connection
// and a reference to the HTTP server serving the request
// see https://github.com/golang/go/issues/16100 for more information
type DeadlineController struct {
	c net.Conn
	s *http.Server
}

// Feed holds information about a RSS feed
type Feed struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// News holds the information to stream about a single item from a RSS feed
type News struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

func (svc *Service) getFeeds(w http.ResponseWriter, r *http.Request) {

}

func (svc *Service) addFeed(w http.ResponseWriter, r *http.Request) {

}

func (svc *Service) streamItems(w http.ResponseWriter, r *http.Request) {

}
