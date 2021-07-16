package svelte

import (
	"fmt"

	"github.com/bfanger/svelte-ssr-go/internal/javascript"
	"rogchap.com/v8go"
)

type Component struct {
	Runtime *javascript.Runtime
	fn      *v8go.Function
}

func Open(js *javascript.Runtime, filename string) (*Component, error) {
	val, err := js.ExecFile(filename)
	if err != nil {
		return nil, err
	}
	obj, err := val.AsObject()
	if err != nil {
		return nil, err
	}
	render, err := obj.Get("render")
	if err != nil {
		return nil, err
	}
	fn, err := render.AsFunction()
	if err != nil {
		return nil, err
	}
	return &Component{Runtime: js, fn: fn}, nil
}
func (c Component) Render() (string, error) {
	result, err := c.fn.Call()
	if err != nil {
		return "", err
	}
	obj, err := result.AsObject()
	if err != nil {
		return "", err
	}
	css, err := obj.Get("css")
	if err != nil {
		return "", err
	}
	cssObj, err := css.AsObject()
	if err != nil {
		return "", err
	}
	style, err := cssObj.Get("code")
	if err != nil {
		return "", err
	}

	html, err := obj.Get("html")
	if err != nil {
		return "", err
	}
	if fmt.Sprint(style) == "" {
		return html.String(), nil
	}
	return fmt.Sprintf("\n<style>%s<style>\n%s\n", style, html), nil

}
