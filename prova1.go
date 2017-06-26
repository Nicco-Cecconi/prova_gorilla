package main

import (
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

type Page struct {
	Title string
	Body  []byte
}

func loadPage(title string) *Page {
	filename := title + ".txt"
	body, _ := ioutil.ReadFile(filename)
	return &Page{Title: title, Body: body}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/"):]
	p := loadPage(title)
	p = &Page{Title: title}
	t, _ := template.ParseFiles("template/home.html")
	t.Execute(w, p)
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	http.ListenAndServe("https://salty-harbor-92838.herokuapp.com/", r)
}
