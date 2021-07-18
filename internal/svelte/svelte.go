package svelte

import (
	"net/http"

	"github.com/bfanger/svelte-ssr-go/internal/javascript"
)

func NewHandler(filename string, debug bool) http.Handler {
	if debug {
		return &DebugHandler{filename}
	}
	js, err := javascript.New() // @todo Reset isolate after X requests (free memory)
	if err != nil {
		return &ErrorResponse{err: err, debug: debug}
	}
	r, err := NewRoute(js, filename, debug)
	if err != nil {
		return &ErrorResponse{err: err, debug: debug}
	}
	return r
}

type DebugHandler struct {
	filename string
}

func (h *DebugHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// @todo Reuse an isolate for X requests?
	js, err := javascript.New()
	if err != nil {
		writeError(w, err, true)
		return
	}
	defer js.Dispose()
	// Create a new route of every request (reloads all scripts)
	r, err := NewRoute(js, h.filename, true)
	if err != nil {
		writeError(w, err, true)
		return
	}
	r.ServeHTTP(w, req)
}
