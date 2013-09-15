package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"fmt"
)

type HTTPListener struct {
	router *mux.Router
	config *Config
}

func NewHTTPListener(config *Config) *HTTPListener {
	router := mux.NewRouter()
	listener := &HTTPListener{ router, config }

	router.HandleFunc("/calls/new", listener.NewCall)

	return listener
}

func (hl *HTTPListener) Listen() error {
	err := http.ListenAndServe(hl.config.HTTPBindTo, hl.router)
	if err != nil {
		return err
	}

	return nil
}

func (hl *HTTPListener) NewCall(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Incoming request to /calls/new")

	es, err := NewESInbound(hl.config)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	
	es.Setup()
}