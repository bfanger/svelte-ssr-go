package javascript

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"rogchap.com/v8go"
)

type Runtime struct {
	Isolate *v8go.Isolate
	Global  *v8go.ObjectTemplate
	Context *v8go.Context
}

func New() (*Runtime, error) {
	iso, err := v8go.NewIsolate()
	if err != nil {
		return nil, errors.Wrap(err, "creating isolate failed")
	}
	global, err := v8go.NewObjectTemplate(iso)
	if err != nil {
		return nil, errors.Wrap(err, "creating global object failed")
	}

	// @todo Add polyfills into global
	ctx, err := v8go.NewContext(iso, global)
	if err != nil {
		return nil, errors.Wrap(err, "creating context failed")
	}
	// @todo one context per page?
	return &Runtime{Isolate: iso, Global: global, Context: ctx}, nil
}

func (r *Runtime) Close() {
	r.Isolate.Dispose()
}

func (r *Runtime) ExecFile(filename string) (*v8go.Value, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}
	return r.Exec(string(b), filename)
}

func (r *Runtime) Exec(code string, origin string) (*v8go.Value, error) {
	val, err := r.Context.RunScript(code, origin)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// 	printfn, err := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
// 		fmt.Printf("%v", info.Args())
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	global.Set("print", printfn)

// 	if err != nil {
// 		panic(err)
// 	}

// 	runScript(ctx, "build/pages/index.js")
// 	val, err := ctx.RunScript(`svelte.render({})`, "main.go")
// 	if err != nil {
// 		uncaught(err)
// 	}
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

// 	val, err := ctx.RunScript(string(b), filename)
// 	if err != nil {
// 		uncaught(err)
// 	}
// 	return val
// }

// func consoleLog(message string) {
// 	log.Println(message)
// }
