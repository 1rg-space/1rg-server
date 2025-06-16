package templates

import (
	"1rg-server/config"
	"embed"
	"html/template"
	"net/http"
)

//go:embed *.tmpl
var content embed.FS

var templates = template.Must(template.ParseFS(content, "*.tmpl"))

// RenderTemplate renders the named template and handles any errors.
// Nothing should be written to the writer after this function.
func RenderTemplate(w http.ResponseWriter, tmpl string, data any) {
	if config.IsProduction {
		err := templates.ExecuteTemplate(w, tmpl+".tmpl", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// In debug mode, reload the template on each request so template changes
	// show up right away
	t := template.Must(template.ParseGlob("templates/*.tmpl"))
	err := t.ExecuteTemplate(w, tmpl+".tmpl", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
