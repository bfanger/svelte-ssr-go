package svelte

import (
	"net/http"

	"github.com/bfanger/svelte-ssr-go/internal/javascript"
)

func NewHandler(js *javascript.Runtime, filename string, debug bool) http.Handler {
	if debug {
		return &DebugHandler{js, filename}
	}
	r, err := NewRoute(js, filename, debug)
	if err != nil {
		return &ErrorResponse{err: err, debug: debug}
	}
	return r
}

type DebugHandler struct {
	js       *javascript.Runtime
	filename string
}

func (h *DebugHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Create a new route of every request (reloads all scripts)
	r, err := NewRoute(h.js, h.filename, true)
	if err != nil {
		writeError(w, err, true)
		return
	}
	r.ServeHTTP(w, req)
}
