package svelte

import (
	"fmt"
	"log"
	"net/http"
)

type ErrorResponse struct {
	debug bool
	err   error
}

func (e *ErrorResponse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeError(w, e.err, e.debug)
}

func errorHandleFunc(err error, debug bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, err, debug)
	}
}

func writeError(w http.ResponseWriter, err error, debug bool) {
	log.Printf("%+v", err)
	w.WriteHeader(500)
	if debug {
		w.Write([]byte(fmt.Sprintf("Error: %+v", err)))
	} else {
		w.Write([]byte("Internal server error"))
	}
}
