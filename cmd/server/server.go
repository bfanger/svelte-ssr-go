package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"flag"

	"github.com/bfanger/svelte-ssr-go/cmd/server/todos"
	"github.com/bfanger/svelte-ssr-go/internal/svelte"
)

func main() {
	port := flag.Int("p", 8080, "Port")
	debug := flag.Bool("d", false, "Debug")
	flag.Parse()

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
	home := svelte.NewHandler("index", *debug)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 Not found")
			return
		}
		home.ServeHTTP(w, r)

	})
	http.Handle("/about", svelte.NewHandler("about", *debug))
	h := &todos.TodoHandler{Page: svelte.NewHandler("todos/index", *debug)}
	h.Add("Test123")

	http.Handle("/todos.json", h)
	http.Handle("/todos/", h)

	http.HandleFunc("/gc", func(w http.ResponseWriter, r *http.Request) {
		runtime.GC()
	})

	fmt.Print("Svelte ")
	if *debug {
		fmt.Print("debug-")
	} else {
		fmt.Print("performance-")
	}
	fmt.Printf("server listening on %d\n", *port)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
