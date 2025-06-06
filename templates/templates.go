package templates

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed *.tmpl
var content embed.FS

var templates = template.Must(template.ParseFS(content, "*.tmpl"))

func RenderTemplate(w http.ResponseWriter, tmpl string, data any) {
	err := templates.ExecuteTemplate(w, tmpl+".tmpl", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
