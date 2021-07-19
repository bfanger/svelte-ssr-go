// Just enough backend for the todos example
// (uses a shared in-memory array for all clients)
package todos

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	Uid       string    `json:"uid"`
	CreatedAt time.Time `json:"created_at"`
	Text      string    `json:"text"`
	Done      bool      `json:"done"`
}
type TodoHandler struct {
	Page  http.Handler
	todos []*Todo
}

func (h *TodoHandler) Add(text string) *Todo {
	todo := &Todo{Uid: uuid.New().String(), Text: text, CreatedAt: time.Now()}
	h.todos = append(h.todos, todo)
	return todo
}
func (h *TodoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := strings.ToUpper(r.FormValue("_method"))
	if method == "" {
		method = r.Method
	}
	var uid string
	var todo *Todo
	match := regexp.MustCompile("/todos/([^/]+)\\.json").FindStringSubmatch(r.URL.Path)
	if len(match) == 2 {
		uid = match[1]
		todo = h.GetByUid(uid)
	}

	if method == http.MethodGet {
		if r.URL.Path == "/todos/" {
			h.Page.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/todos.json" {
			w.Write(h.GetAll())
			return
		}
		if todo != nil {
			w.Write(toJson(todo))
			return
		}
	}
	if method == http.MethodPost {
		if r.URL.Path == "/todos.json" {
			todo := h.Add(r.PostFormValue("text"))
			w.Write(toJson(todo))
			return
		}
	}
	if method == http.MethodDelete {
		if uid != "" {
			for i, todo := range h.todos {
				if todo.Uid == uid {
					h.todos = append(h.todos[:i], h.todos[i+1:]...)
					w.Write([]byte("OK"))
					return
				}
			}
		}
	}
	if method == http.MethodPatch && todo != nil {
		r.ParseForm()
		if len(r.PostForm["text"]) != 0 {
			todo.Text = r.PostForm["text"][0]
		}
		if len(r.PostForm["done"]) != 0 {
			todo.Done = r.PostForm["done"][0] == "true"
		}
		w.Write(toJson(todo))
		return
	}
	w.WriteHeader(400)
}

func (h *TodoHandler) GetAll() []byte {
	if h.todos == nil {
		return []byte("[]")
	}
	return toJson(h.todos)
}

func toJson(data interface{}) []byte {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return b
}
func (h *TodoHandler) GetByUid(uid string) *Todo {
	for _, todo := range h.todos {
		if todo.Uid == uid {
			return todo
		}
	}
	return nil

}
