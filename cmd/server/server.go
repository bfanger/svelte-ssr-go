package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bfanger/svelte-ssr-go/internal/svelte"
)

func main() {
	const debug = true
	const port = ":8080"

	http.Handle("/", svelte.NewHandler("build/routes/index.js", debug))
	http.Handle("/about", svelte.NewHandler("build/routes/about.js", debug))

	fmt.Printf("Server started on %s\n", port)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
