package javascript

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/bfanger/svelte-ssr-go/internal/util"
	"github.com/pkg/errors"
	"rogchap.com/v8go"
)

type Runtime struct {
	Isolate *v8go.Isolate
	Context *v8go.Context
	Fetch   *v8go.Function
}

func New() (*Runtime, error) {
	iso, err := v8go.NewIsolate()
	if err != nil {
		return nil, errors.Wrap(err, "creating isolate failed")
	}
	ctx, err := v8go.NewContext(iso)
	if err != nil {
		return nil, errors.Wrap(err, "creating context failed")
	}
	fetch, err := v8go.NewFunctionTemplate(iso, fetchPolyfill)
	if err != nil {
		return nil, errors.Wrap(err, "creating fetch polyfill failed")
	}
	return &Runtime{Isolate: iso, Context: ctx, Fetch: fetch.GetFunction(ctx)}, nil
}

func (r *Runtime) Dispose() {
	r.Context.Close()
	r.Isolate.Dispose()
}

func (r *Runtime) ExecFile(filename string) (*v8go.Value, error) {
	b, err := os.ReadFile(filename)
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

func (r *Runtime) NewObject() (*v8go.Object, error) {
	empty, err := r.Context.RunScript("({});", "javascript.go")
	if err != nil {
		return nil, errors.Wrap(err, "creating empty object failed")
	}
	return empty.AsObject()
}

func (r *Runtime) ErrorValue(message error) *v8go.Value {
	return ErrorValue(r.Context, message)
}

// Create an javascript error
func ErrorValue(ctx *v8go.Context, message error) *v8go.Value {
	json, err := json.Marshal(message.Error())
	if err != nil {
		panic(err)
	}
	val, err := ctx.RunScript(fmt.Sprintf("(new Error(%s))", json), "javascript.go")
	if err != nil {
		panic(err)
	}
	return val
}
func (r *Runtime) PrintJSON(val *v8go.Value) {
	json, err := v8go.JSONStringify(r.Context, val)
	if err != nil {
		fmt.Printf("%#v\n", val)
	} else {
		fmt.Printf("%s\n", json)
	}
}

// 	printfn, err := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
// 		fmt.Printf("%v", info.Args())
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	global.Set("print", printfn)

// json, err := v8go.JSONStringify(ctx, val)
// if err != nil {
// 	fmt.Printf("Result: %+v\n", val)
// } else {
// 	fmt.Printf("Result: %s\n", json)
// }

// func uncaught(err error) {
// 	e := err.(*v8go.JSError)
// 	fmt.Println(e.Message)
// 	fmt.Println(e.Location)
// 	fmt.Println(e.StackTrace)

// 	log.Fatalf("Uncaught Error: %+v\n", err)
// }

// func consoleLog(message string) {
// 	log.Println(message)
// }

func PropAsObject(o *v8go.Object, key string) (*v8go.Object, error) {
	prop, err := o.Get(key)
	if err != nil {
		return nil, err
	}
	propObj, err := prop.AsObject()
	if err != nil {
		return nil, err
	}
	return propObj, nil
}

func PropAsFunction(o *v8go.Object, key string) (*v8go.Function, error) {
	prop, err := o.Get(key)
	if err != nil {
		return nil, err
	}
	fn, err := prop.AsFunction()
	if err != nil {
		return nil, err
	}
	return fn, nil
}

func WaitFor(p *v8go.Promise) (v8go.PromiseState, *v8go.Value) {
	c := make(chan v8go.PromiseState)
	go func() {
		if p.State() != v8go.Pending {
			c <- p.State()
			return
		}
		p.Then(func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			c <- v8go.Fulfilled
			return p.Result()
		})
		p.Catch(func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			c <- v8go.Rejected
			return p.Result()
		})
	}()
	return <-c, p.Result()
}

type FetchResponse struct {
	Ok bool `json:"ok"`
}

func fetchPolyfill(info *v8go.FunctionCallbackInfo) *v8go.Value {
	ctx := info.Context()
	deferred, err := v8go.NewPromiseResolver(ctx)
	util.AssertNoError(err)
	args := info.Args()
	if len(args) == 1 && args[0].IsString() {
		url := args[0].String()
		if strings.HasPrefix(url, "/") {
			url = "http://localhost:8080" + url
			// @todo detect or configure host
		}
		r, err := http.Get(url)
		if err != nil {
			deferred.Reject(ErrorValue(ctx, err))
			return deferred.Value
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			deferred.Reject(ErrorValue(ctx, err))
			return deferred.Value
		}
		ok := "true"
		if r.StatusCode != http.StatusOK {
			ok = "false"
		}
		script :=
			fmt.Sprintf(`({
	ok: %s,
	async json() {
		return %s
	}})`, ok, string(b))

		data, err := ctx.RunScript(script, "javascript.go")
		if err != nil {
			deferred.Reject(ErrorValue(ctx, err))
			return deferred.Value
		}
		deferred.Resolve(data)
	}
	return deferred.Value
}
