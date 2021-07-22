package javascript

import (
	"os"

	"github.com/pkg/errors"
	"rogchap.com/v8go"
)

type Runtime struct {
	Isolate *v8go.Isolate
	Context *v8go.Context
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
	return &Runtime{Isolate: iso, Context: ctx}, nil
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
