package svelte

import (
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
func (r *Route) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	var result *Result
	var err error
	if r.Layout != nil {
		result, err = r.Layout.Render(r.Layout.DefaultProps, r.LayoutOptions)
		if err != nil {
			writeError(w, err, r.debug)
			return
		}
	} else {
		result, err = r.Component.Render()
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
