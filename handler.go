package jparser

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/jsonapi"
)

// RequestHandler handles http requests
type RequestHandler struct {
	timer func() <-chan time.Time
}

// NewRequestHandler creates new http request handler
func NewRequestHandler() *RequestHandler {
	return &RequestHandler{
		timer: func() <-chan time.Time {
			t := rand.Intn(1000) + 500
			// wait random time between 0.5 and 1.5 second
			return time.After(time.Millisecond * time.Duration(t))
		},
	}
}

// DoSomeWork does some dummy work
func (h *RequestHandler) DoSomeWork(w http.ResponseWriter, r *http.Request) {
	t := h.timer()
	w.WriteHeader(http.StatusOK)
	<-t
}

// ParseJSON parses and beautifies input in json format
func (h *RequestHandler) ParseJSON(w http.ResponseWriter, r *http.Request) {
	if len(r.Header) > 0 && r.Header["Content-Type"][0] != "application/json" {
		http.Error(w, "wrong content type", http.StatusBadRequest)
		return
	}

	t := h.timer()

	var dat interface{}

	// parse request
	dc := json.NewDecoder(r.Body)
	if err := dc.Decode(&dat); err != nil {
		log.Println(err)
		http.Error(w, "json parsing error", http.StatusInternalServerError)
		return
	}

	// format json
	resp, err := json.MarshalIndent(dat, "", "  ")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)

	<-t
}

// ParseJSONapi parses and beautifies input in JSON:API format
func (h *RequestHandler) ParseJSONapi(w http.ResponseWriter, r *http.Request) {
	if len(r.Header) > 0 && r.Header["Content-Type"][0] != jsonapi.MediaType {
		http.Error(w, "wrong content type", http.StatusBadRequest)
		return
	}

	t := h.timer()

	var dummy jsonapi.OnePayload
	var buf bytes.Buffer

	tee := io.TeeReader(r.Body, &buf)

	// validate JSON:API request
	err := jsonapi.UnmarshalPayload(tee, &dummy)
	if err != nil {
		log.Println(err)
		http.Error(w, "json parsing error", http.StatusInternalServerError)
		return
	}

	var dat interface{}

	// parse request
	dc := json.NewDecoder(&buf)
	if err := dc.Decode(&dat); err != nil {
		log.Println(err)
		http.Error(w, "json parsing error", http.StatusInternalServerError)
		return
	}

	// format json
	resp, err := json.MarshalIndent(dat, "", "  ")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)

	<-t
}
