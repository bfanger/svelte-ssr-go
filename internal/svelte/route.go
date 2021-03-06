package svelte

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/bfanger/svelte-ssr-go/internal/javascript"
	"rogchap.com/v8go"
)

type Route struct {
	js            *javascript.Runtime
	debug         bool
	InlineCSS     string
	InlineScript  string
	Template      *template.Template
	Component     *Component
	Layout        *Component
	LayoutOptions *v8go.Object
}

func NewRoute(js *javascript.Runtime, filename string, debug bool) (*Route, error) {
	r := &Route{js: js, debug: debug}
	var err error
	r.Component, err = NewComponent(js, filename)
	if err != nil {
		return nil, err
	}
	if r.Component.CSSFile != "" {
		// @todo Use external (hashed) url in a <link>?
		css, err := os.ReadFile(r.Component.CSSFile)
		if err != nil {
			return nil, err
		}
		r.InlineCSS = string(css)
	}
	if r.Component.JsClientFile != "" {
		js, err := os.ReadFile(r.Component.JsClientFile)
		if err != nil {
			return nil, err
		}
		// @todo Use external (hashed) url in a <script src>?
		r.InlineScript = "var componentModule = " + string(js)
	}

	layoutPath := regexp.MustCompile("[^/]+$").ReplaceAllString(filename, "__layout")
	stat, _ := os.Stat("build/server/" + layoutPath + ".js")
	if stat == nil {
		// @todo Traverse folders inside build/routes
		layoutPath = strings.TrimLeft(path.Dir(path.Dir(layoutPath))+"/__layout", "./")
		stat, _ = os.Stat("build/server/" + layoutPath + ".js")
	}
	if stat != nil {
		r.Layout, err = NewComponent(js, layoutPath)
		if err != nil {
			return nil, err
		}
		// @todo props from load
		slots, err := r.Component.js.Context.RunScript(`
(function () {
	return {
		$$slots: {
			default() { return this.__go_Component_render().html; }
		}
	}
})();`, "route.go")
		r.LayoutOptions, err = slots.AsObject()
		if err != nil {
			return nil, err
		}

		definedSlots, err := javascript.PropAsObject(r.LayoutOptions, "$$slots")
		if err != nil {
			return nil, err
		}
		definedSlots.Set("__go_Component_render", r.Component.render)
		if r.Layout.CSSFile != "" {
			css, err := os.ReadFile(r.Layout.CSSFile)
			if err != nil {
				return nil, err
			}
			// @todo Use external (hashed) url in a <link>?
			r.InlineCSS += string(css)
		}
		if r.Layout.JsClientFile != "" {
			js, err := os.ReadFile(r.Layout.JsClientFile)
			if err != nil {
				return nil, err
			}
			// @todo Use external (hashed) url in a <script src>?
			r.InlineScript += "\nvar layoutModule = " + string(js) + ";"
			r.InlineScript += `
var slot = new componentModule.default({});
var app = new layoutModule.default({ target: document.getElementById("svelte"), hydrate: true, props: {
  $$scope: {ctx: slot.ctx},
  $$slots:{ default: [ () => slot.$$.fragment ] },
}});
`
		} else {
			r.InlineScript += "\nvar app = new componentModule.default({ target: document.getElementById('svelte'), hydrate: true });"
		}
	}

	r.Template, err = appHTML()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Route) Load(c *Component) *Payload {
	p := &Payload{Status: http.StatusOK, Props: c.DefaultProps}
	if c.Load == nil {
		return p
	}
	context, err := c.js.NewObject()
	if err != nil {
		return &Payload{Status: http.StatusInternalServerError, Err: err}
	}
	context.Set("fetch", c.js.Fetch)
	val, err := c.Load.Call(context)
	if err != nil {
		return &Payload{Status: http.StatusInternalServerError, Err: err}
	}
	promise, err := val.AsPromise()
	if err != nil {
		return &Payload{Status: http.StatusInternalServerError, Err: err}
	}
	state, value := javascript.WaitFor(promise)
	props, err := value.AsObject()
	if err != nil {
		return &Payload{Status: http.StatusInternalServerError, Err: err}
	}
	if state == v8go.Rejected {
		message := "promise rejected"
		if err == nil {
			jsMessage, err := props.Get("message")
			fmt.Println(err)
			if err == nil && jsMessage.IsString() {
				message = jsMessage.String()
			}
		}
		return &Payload{Status: http.StatusInternalServerError, Props: c.DefaultProps, Err: errors.New(message)}
	}

	return &Payload{Status: http.StatusOK, Props: props}
}

func (r *Route) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	p := r.Load(r.Component)
	if p.Err != nil {
		writeError(w, p.Err, r.debug)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	var result *Result
	var err error
	if r.Layout != nil {
		// @todo Use the result of the Load data
		result, err = r.Layout.Render(r.Layout.DefaultProps, r.LayoutOptions)
		if err != nil {
			writeError(w, err, r.debug)
			return
		}
	} else {
		result, err = r.Component.Render(p.Props)
		if err != nil {
			writeError(w, err, r.debug)
			return
		}
	}
	result.Head += "<style>\n" + r.InlineCSS + "</style>\n"
	result.HTML += "\n<script>\n" + r.InlineScript + "</script>\n"
	if r.debug {
		result.HTML += "<script src=\"http://localhost:35729/livereload.js\"></script>\n"
	}
	err = r.Template.Execute(w, result)
	if err != nil {
		writeError(w, err, r.debug)
	}
}

func appHTML() (*template.Template, error) {
	app, err := os.ReadFile("example/src/app.html")
	if err != nil {
		return nil, err
	}
	html := string(app)
	html = regexp.MustCompile("%svelte.head%").ReplaceAllString(html, "{{ .Head }}")
	html = regexp.MustCompile("%svelte.body%").ReplaceAllString(html, "{{ .HTML }}")
	t, err := template.New("app.html").Parse(html)
	if err != nil {
		return nil, err
	}
	return t, nil
}

type Payload struct {
	Props  *v8go.Object
	Err    error
	Status int
}

func (p *Payload) PropsAsJson(ctx *v8go.Context) string {
	json, err := v8go.JSONStringify(ctx, p.Props)
	if err != nil {
		return "{}"
	}
	return json
}
