package main

import (
	"embed"
	"log"
	"net/http"
	"text/template"
)

//go:embed ../../web/templates/*html
//go:embed  ../../web/static/*
var fs embed.FS

var tpl = template.Must(template.ParseFS(fs, "web/templates/*.html"))

func main() {
	http.Handle("/static/", http.FileServer(http.FS(fs)))
	http.HandleFunc("/", handleHome)

	log.Println("Server starting on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}

func handleHome(w http.ResponseWriter, r *http.Request) {
	err := tpl.Execute(w, "index")
	if err != nil {
		http.Error(w, "error parsing index template html", http.StatusInternalServerError)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "error parsing html template", http.StatusInternalServerError)
	}
	t.Execute(w, data)
}
