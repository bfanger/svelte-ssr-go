package svelte

import (
	"net/http"
	"os"
	"regexp"
	"text/template"

	"github.com/bfanger/svelte-ssr-go/internal/javascript"
	"rogchap.com/v8go"
)

type Route struct {
	js            *javascript.Runtime
	debug         bool
	InlineCss     string
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
	layoutPath := regexp.MustCompile("[^/]+$").ReplaceAllString(filename, "__layout.js")
	stat, _ := os.Stat(layoutPath)
	if stat == nil {
		r.Layout = nil
	} else {
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
	}

	r.InlineCss = ""
	if r.Layout.CssFile != "" {
		css, err := os.ReadFile(r.Layout.CssFile)
		if err != nil {
			return nil, err
		}
		// @todo Use external (hashed) url in a <link>?
		r.InlineCss += "<style>\n" + string(css) + "</style>\n"
	}
	if r.Component.CssFile != "" {
		css, err := os.ReadFile(r.Component.CssFile)
		if err != nil {
			return nil, err
		}
		// @todo Use external (hashed) url in a <link>?
		r.InlineCss += "<style>\n" + string(css) + "</style>\n"
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
	result.Head += r.InlineCss
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
