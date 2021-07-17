package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bfanger/svelte-ssr-go/internal/javascript"
	"github.com/bfanger/svelte-ssr-go/internal/svelte"
)

func main() {
	const debug = true
	const port = ":8080"

	js, err := javascript.New() // @todo Reuse a vm or use a pool?
	if err != nil {
		panic(err)
	}

	fs := http.FileServer(http.Dir("example/static"))
	dir, err := os.ReadDir("example/static")
	if err == nil {
		for _, entry := range dir {
			if entry.IsDir() {
				http.Handle("/"+entry.Name()+"/*", fs)
			} else {
				http.Handle("/"+entry.Name(), fs)
			}
		}
	}

	// @todo crawl routes folder
	http.Handle("/", svelte.NewHandler(js, "build/routes/index.js", debug))
	http.Handle("/about", svelte.NewHandler(js, "build/routes/about.js", debug))

	http.HandleFunc("/gc", func(w http.ResponseWriter, r *http.Request) {

		stats := js.Isolate.GetHeapStatistics()
		fmt.Fprintf(w, "<html><pre style=\"white-space: pre-line;\"'>%+v</pre></html>", stats)
	})

	fmt.Print("Svelte ")
	if debug {
		fmt.Print("debug-")
	} else {
		fmt.Print("production-")
	}
	fmt.Printf("server listening on %s\n", port)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
