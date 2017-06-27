package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"

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
	port := strconv.FormatInt(8080, 10)
	fmt.Print("Serving and listening at port 8080")
	http.ListenAndServe("10.0.2.15:"+port, r)
}
