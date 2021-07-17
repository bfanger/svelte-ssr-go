package svelte

import (
	"net/http"

	"github.com/bfanger/svelte-ssr-go/internal/javascript"
)

func NewHandler(filename string, debug bool) http.Handler {
	js, err := javascript.New() // @todo Reuse a vm?
	if err != nil {
		return &ErrorResponse{err: err, debug: debug}
	}
	r, err := NewRoute(js, filename, debug)
	if err != nil {
		return &ErrorResponse{err: err, debug: debug}
	}
	return r
}

// A handeFunc
// func NewHandlerFuncProduction(filename string) func(http.ResponseWriter, *http.Request) {

// 	return func(w http.ResponseWriter, r *http.Request) {
// 		output, err := c.Handle(w, r)
// 		if err != nil {
// 			handleError(w, err, false)
// 			return
// 		}
// 		if css != nil {
// 			w.Write([]byte("<style>"))
// 			w.Write(css)
// 			w.Write([]byte("</style>"))
// 		}
// 		// return fmt.Sprintf("\n<style>%s</style>\n%s\n", r.Css, r.Html)

// 		w.Write([]byte(output.Head))
// 		w.Write(output.Bytes())
// 	}
// }
