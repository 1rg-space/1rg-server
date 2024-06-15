package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed assets
//go:embed templates
var content embed.FS

var templates = template.Must(template.ParseFS(content, "templates/*"))

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "index", nil)
	})
	http.Handle("GET /assets/", http.FileServerFS(content))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
