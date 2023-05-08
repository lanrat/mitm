// Package server handles the http server for the frontend
package server

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
)

const timeoutDuration = 15 * time.Second

// Server struct for holding server resources
type Server struct {
	router     *mux.Router
	listenAddr string
	//Session    sessions.Store
}

// New creates a new server object with the default (included) handlers
func New(listenAddr string) (*Server, error) {
	server := &Server{
		listenAddr: listenAddr,
		//Session:    sessions.NewFilesystemStore("", []byte("something-very-secret")),
	}

	// setup server
	server.router = mux.NewRouter() //.StrictSlash(true)

	// serve static content
	//static := http.FileServer(http.Dir("data"))
	//server.router.PathPrefix("/").Handler(static)

	// setup robots.txt
	// server.router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "static/robots.txt")
	// }).Methods(http.MethodGet)

	return server, nil
}

// Get registers a HTTP GET to the router & handler
func (s *Server) Get(path string, fn http.HandlerFunc) {
	s.router.Handle(path, fn).Methods(http.MethodGet)
}

// GetPrefix registers a HTTP GET for the prefix to the router & handler
func (s *Server) GetPrefix(path string, fn http.HandlerFunc) {
	s.router.PathPrefix(path).Handler(fn)
}

// Post registers a HTTP POST to the router & handler
func (s *Server) Post(path string, fn http.HandlerFunc) {
	s.router.Handle(path, fn).Methods(http.MethodPost)
}

// Start Starts the server, blocking function
func (s *Server) Start() error {
	// prep proxy handler
	h := handlers.ProxyHeaders(s.router)
	// setup logging
	h = handlers.CustomLoggingHandler(os.Stdout, h, writeLog)
	// add recovery
	h = handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(h)
	// timeouts
	h = http.TimeoutHandler(h, timeoutDuration, "Timeout!")

	var group errgroup.Group
	group.Go(func() error {
		// run server
		srv := &http.Server{
			Handler:      h,
			Addr:         net.JoinHostPort(s.listenAddr, "80"),
			WriteTimeout: timeoutDuration,
			ReadTimeout:  timeoutDuration,
		}
		return srv.ListenAndServe()

	})
	group.Go(func() error {
		// run server
		srv := &http.Server{
			Handler:      h,
			Addr:         net.JoinHostPort(s.listenAddr, "443"),
			WriteTimeout: timeoutDuration,
			ReadTimeout:  timeoutDuration,
			TLSConfig:    getTLSConfig(),
		}
		return srv.ListenAndServeTLS("", "")
	})
	return group.Wait()
}
