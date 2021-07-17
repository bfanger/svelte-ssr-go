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
	index, err := svelte.NewComponent(js, "build/routes/index.js")
	util.AssertNoError(err)
	indexOutput, err := index.Render()
	util.AssertNoError(err)
	fmt.Printf("index.svelte: %+v\n", indexOutput)

	about, err := svelte.NewComponent(js, "build/routes/about.js")
	util.AssertNoError(err)
	aboutOutput, err := about.Render()
	util.AssertNoError(err)
	fmt.Printf("about.svelte: %+v\n", aboutOutput)
}
