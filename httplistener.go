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

	router.HandleFunc("/calls/new", listener.NewCall).Methods("POST")

	return listener
}

func (hl *HTTPListener) Listen() error {
	err := http.ListenAndServe(hl.config.HTTPBindTo, hl.router)
	if err != nil {
		return err
	}

	return nil
}

func (hl *HTTPListener) WriteError(w http.ResponseWriter, error string) {
	w.WriteHeader(500)
	w.Write([]byte(error))
}

func (hl *HTTPListener) NewCall(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println("Error:", err)
		hl.WriteError(w, "Failed to create outbound call")
		return
	}

	from := r.Form.Get("from")
	to := r.Form.Get("to")

	es, err := NewESInbound(hl.config)
	if err != nil {
		fmt.Println("Error:", err)
		hl.WriteError(w, "Failed to create outbound call")
		return
	}
	
	es.Setup()
	es.Originate(from, to)
	es.Close()

	w.Write([]byte("Success"))
}