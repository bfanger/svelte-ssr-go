package svelte

import (
	"net/http"
	"os"
	"regexp"
	"text/template"

	"github.com/bfanger/svelte-ssr-go/internal/javascript"
)

type Route struct {
	js        *javascript.Runtime
	filename  string
	debug     bool // Reload component & app.html on every request + render error
	InlineCss string
	Template  *template.Template
	Component *Component
}

func NewRoute(js *javascript.Runtime, filename string, debug bool) (*Route, error) {
	r := &Route{js: js, filename: filename, debug: debug}
	if debug == false {
		if err := r.Reload(); err != nil {
			return nil, err
		}
	}

	return r, nil
}
func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.debug {
		if err := r.Reload(); err != nil {
			writeError(w, err, r.debug)
			return
		}
	}
	result, err := r.Component.Render()
	if err != nil {
		writeError(w, err, r.debug)
		return
	}
	result.Head += r.InlineCss
	err = r.Template.Execute(w, result)
	if err != nil {
		writeError(w, err, r.debug)
	}
}

func (r *Route) Reload() error {
	var err error
	r.Component, err = NewComponent(r.js, r.filename)
	if err != nil {
		return err
	}
	r.InlineCss = ""
	if r.Component.CssFile != "" {
		css, err := os.ReadFile(r.Component.CssFile)
		if err != nil {
			return err
		}
		// @todo Use external (hashed) url in a <link>?
		r.InlineCss = "<style>" + string(css) + "</style>"
	}
	r.Template, err = appHtml()
	if err != nil {
		return err
	}
	return nil
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
