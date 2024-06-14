package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "pages/index.html")
	})
	http.Handle("GET /assets/", http.FileServer(http.Dir(".")))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
