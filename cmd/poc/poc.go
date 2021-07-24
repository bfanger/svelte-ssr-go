package main

import (
	"fmt"

	"github.com/bfanger/svelte-ssr-go/internal/javascript"
	"github.com/bfanger/svelte-ssr-go/internal/svelte"
	"github.com/bfanger/svelte-ssr-go/internal/util"
)

func main() {
	js, err := javascript.New()
	util.AssertNoError(err)
	defer js.Dispose()

	renderRoute(js, "index")
	// renderRoute(js, "about")
	// renderRoute(js, "todos/index")
}

func renderRoute(js *javascript.Runtime, filename string) {
	r, err := svelte.NewRoute(js, filename, true)
	util.AssertNoError(err)
	p := r.Load(r.Component)
	util.AssertNoError(p.Err)

	fmt.Printf("%#v\n\n", p)
	js.PrintJSON(p.Props.Value)
	output, err := r.Component.Render(p.Props)
	util.AssertNoError(err)
	fmt.Printf("%s.svelte: %+v\n", filename, output)
}
