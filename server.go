package jparser

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/didip/tollbooth"
	"github.com/gorilla/mux"
)

// Handler is a request handler interface
type Handler interface {
	DoSomeWork(http.ResponseWriter, *http.Request)
	ParseJSON(http.ResponseWriter, *http.Request)
	ParseJSONapi(http.ResponseWriter, *http.Request)
}

// Server is an main application server
type Server struct {
	// request connection limit per second
	reqLimit float64
	handler  Handler
}

// NewServer creates new server object
//
// reqLimit is number of connections per second allowed for each ip
func NewServer(reqLimit float64) *Server {
	return &Server{
		reqLimit: reqLimit,
		handler:  NewRequestHandler(),
	}
}

// Start runs the server
func (s *Server) Start() error {

	r := s.getRouter()

	s.handleInterrupt()

	// start server
	log.Println("starting server")
	return http.ListenAndServe(":8000", r)
}

// Stop shuts the server down
func (s *Server) Stop() {
	fmt.Println()
	log.Println("shutdown server")
}

func (s *Server) getRouter() *mux.Router {
	r := mux.NewRouter()
	lim := tollbooth.NewLimiter(s.reqLimit, nil)
	// setup uri handlers
	r.Handle("/worker", tollbooth.LimitFuncHandler(lim, s.handler.DoSomeWork)).Methods("GET")
	r.Handle("/json-parser", tollbooth.LimitFuncHandler(lim, s.handler.ParseJSON)).Methods("POST")
	r.Handle("/jsonapi-parser", tollbooth.LimitFuncHandler(lim, s.handler.ParseJSONapi)).Methods("POST")
	return r
}

func (s *Server) handleInterrupt() {
	var signalChan = make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-signalChan
		s.Stop()
		os.Exit(0)
	}()
}
