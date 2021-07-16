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
	index, err := svelte.Open(js, "build/pages/index.js")
	util.AssertNoError(err)
	val, err := index.Render()
	util.AssertNoError(err)
	fmt.Println(val)

	about, err := svelte.Open(js, "build/pages/about.js")
	util.AssertNoError(err)
	val, err = about.Render()
	util.AssertNoError(err)
	fmt.Println(val)

	// 	printfn, err := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
	// 		fmt.Printf("%v", info.Args())
	// 		return nil
	// 	})
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	global.Set("print", printfn)

	// 	obj, err := val.AsObject()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	css, err := obj.Get("css")
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	cssObj, err := css.AsObject()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	style, err := cssObj.Get("code")
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	html, err := obj.Get("html")
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Printf("\n<style>%s<style>\n%s\n", style, html)

	// 	// json, err := v8go.JSONStringify(ctx, val)
	// 	// if err != nil {
	// 	// 	fmt.Printf("Result: %+v\n", val)
	// 	// } else {
	// 	// 	fmt.Printf("Result: %s\n", json)
	// 	// }
	// }

	// func readFile(filename string) string {
	// 	f, err := os.Open(filename)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	b, err := io.ReadAll(f)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	return string(b)
	// }

	// func uncaught(err error) {
	// 	e := err.(*v8go.JSError)
	// 	fmt.Println(e.Message)
	// 	fmt.Println(e.Location)
	// 	fmt.Println(e.StackTrace)

	// 	log.Fatalf("Uncaught Error: %+v\n", err)

	// 	// panic(err)
	// }

	// func runScript(ctx *v8go.Context, filename string) *v8go.Value {
	// 	f, err := os.Open(filename)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	b, err := io.ReadAll(f)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	val, err := ctx.RunScript(string(b), filename)
	// 	if err != nil {
	// 		uncaught(err)
	// 	}
	// 	return val
}

// func consoleLog(message string) {
// 	log.Println(message)
// }
