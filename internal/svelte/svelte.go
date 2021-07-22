package svelte

import (
	"net/http"
	"sync"

	"github.com/bfanger/svelte-ssr-go/internal/javascript"
)

func NewHandler(filename string, debug bool) http.Handler {
	if debug {
		return &DebugHandler{filename}
	}
	return &OptimizedHandler{Filename: filename, ReloadAfter: 500}
}

// DebugHandler: Use a fresh VM and reloads all scripts for every request
type DebugHandler struct {
	filename string
}

func (h *DebugHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	js, err := javascript.New()
	if err != nil {
		writeError(w, err, true)
		return
	}
	defer js.Dispose()
	r, err := NewRoute(js, h.filename, true)
	if err != nil {
		writeError(w, err, true)
		return
	}
	r.ServeHTTP(w, req)
}

// OptimizedHandler: Re-uses a v8go.Context and components until the `ReloadAfter` is reached.
// v8go leaks memory (by-design) Disposing the Context and Isolate keeps memory usage acceptable
type OptimizedHandler struct {
	Filename    string
	ReloadAfter int32
	processor   *Processor
	countdown   int32
	mutex       sync.Mutex
}
type Processor struct {
	route http.Handler
	js    *javascript.Runtime
	wg    *sync.WaitGroup
}

func (h *OptimizedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, err := h.AutoReload()
	if err != nil {
		writeError(w, err, false)
		return
	}
	defer p.wg.Done()
	p.route.ServeHTTP(w, r)
}

func (h *OptimizedHandler) AutoReload() (*Processor, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.countdown == 0 {
		if h.processor != nil {
			p := h.processor
			go func() {
				p.wg.Wait()    // Wait for active requests to finish
				p.js.Dispose() // Free memory
			}()
		}
		js, err := javascript.New() // @todo Share isolate between routes?
		if err != nil {
			return nil, err
		}
		route, err := NewRoute(js, h.Filename, false)
		if err != nil {
			return nil, err
		}
		h.processor = &Processor{route: route, js: js, wg: &sync.WaitGroup{}}
		h.countdown = h.ReloadAfter + 1

	}
	if h.ReloadAfter != 0 {
		h.countdown--
	}
	h.processor.wg.Add(1)
	return h.processor, nil

}
