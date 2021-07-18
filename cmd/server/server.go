package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"flag"

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
	http.Handle("/", svelte.NewHandler("build/routes/index", *debug))
	http.Handle("/about", svelte.NewHandler("build/routes/about", *debug))

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
