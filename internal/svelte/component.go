package svelte

import (
	"os"

	"github.com/bfanger/svelte-ssr-go/internal/javascript"
	"rogchap.com/v8go"
)

type Component struct {
	js           *javascript.Runtime
	DefaultProps *v8go.Object
	CSSFile      string
	JsClientFile string
	component    *v8go.Object
	render       *v8go.Function
}

func NewComponent(js *javascript.Runtime, filename string) (*Component, error) {
	result, err := js.ExecFile("build/server/" + filename + ".js")
	if err != nil {
		return nil, err
	}
	exports, err := result.AsObject()
	if err != nil {
		return nil, err
	}
	defaultExport, err := javascript.PropAsObject(exports, "default")
	if err != nil {
		return nil, err
	}
	render, err := javascript.PropAsFunction(defaultExport, "render")
	if err != nil {
		return nil, err
	}
	css := "build/client/" + filename + ".css"
	stat, _ := os.Stat(css)
	if stat == nil {
		css = ""
	}
	client := "build/client/" + filename + ".js"
	stat, _ = os.Stat(client)
	if stat == nil {
		client = ""
	}
	props, err := js.NewObject()
	if err != nil {
		return nil, err
	}
	return &Component{js: js, DefaultProps: props, CSSFile: css, JsClientFile: client, component: defaultExport, render: render}, nil
}

type Result struct {
	HTML string
	Head string
	CSS  string
}

func (c Component) Render(args ...v8go.Valuer) (*Result, error) {
	result, err := c.render.Call(args...)
	if err != nil {
		return nil, err
	}
	obj, err := result.AsObject()
	if err != nil {
		return nil, err
	}

	// json, err := v8go.JSONStringify(c.js.Context, obj)
	// if err != nil {
	// 	fmt.Printf("Result: %+v\n", obj)
	// } else {
	// 	fmt.Printf("Result: %s\n", json)
	// }

	css, err := javascript.PropAsObject(obj, "css")
	if err != nil {
		return nil, err
	}
	cssCode, err := css.Get("code")
	if err != nil {
		return nil, err
	}

	head, err := obj.Get("head")
	if err != nil {
		return nil, err
	}

	html, err := obj.Get("html")
	if err != nil {
		return nil, err
	}
	return &Result{
		Head: head.String(),
		CSS:  cssCode.String(),
		HTML: html.String(),
	}, nil
}
