package svelte

import (
	"net/http"
	"os"
	"path"
	"regexp"
	"text/template"

	"github.com/bfanger/svelte-ssr-go/internal/javascript"
	"rogchap.com/v8go"
)

type Route struct {
	js            *javascript.Runtime
	debug         bool
	InlineCss     string
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
	if r.Component.CssFile != "" {
		// @todo Use external (hashed) url in a <link>?
		css, err := os.ReadFile(r.Component.CssFile)
		if err != nil {
			return nil, err
		}
		r.InlineCss = string(css)
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
	stat, _ := os.Stat(layoutPath + ".server.js")
	if stat == nil {
		// @todo Traverse folders inside build/routes
		layoutPath = path.Dir(path.Dir(layoutPath)) + "/__layout"
		stat, _ = os.Stat(layoutPath + ".server.js")
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
		if r.Layout.CssFile != "" {
			css, err := os.ReadFile(r.Layout.CssFile)
			if err != nil {
				return nil, err
			}
			// @todo Use external (hashed) url in a <link>?
			r.InlineCss += string(css)
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

	r.Template, err = appHtml()
	if err != nil {
		return nil, err
	}
	return r, nil
}
func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	var result *Result
	var err error
	if r.Layout != nil {
		result, err = r.Layout.Render(r.js.EmptyObject, r.LayoutOptions)
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
	result.Head += "<style>\n" + r.InlineCss + "</style>\n"
	result.Html += "\n<script>\n" + r.InlineScript + "</script>\n"
	err = r.Template.Execute(w, result)
	if err != nil {
		writeError(w, err, r.debug)
	}
}

func appHtml() (*template.Template, error) {
	app, err := os.ReadFile("example/src/app.html")
	if err != nil {
		return nil, err
	}
	html := string(app)
	html = regexp.MustCompile("%svelte.head%").ReplaceAllString(html, "{{ .Head }}")
	html = regexp.MustCompile("%svelte.body%").ReplaceAllString(html, "{{ .Html }}")
	t, err := template.New("app.html").Parse(html)
	if err != nil {
		return nil, err
	}
	return t, nil
}
