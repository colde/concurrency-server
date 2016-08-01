package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
)

var adminpw = "foobar"
var addr = flag.String("addr", ":8026", "http service address")

var indexTemplate = template.Must(template.ParseFiles("static/index.html"))
var playerTemplate = template.Must(template.ParseFiles("static/player.html"))

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	switch r.URL.Path {
	case "/":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		indexTemplate.Execute(w, r.Host)
	case "/player":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		playerTemplate.Execute(w, r.Host)
	default:
		http.Error(w, "Not found", 404)
		return
	}
	if r.URL.Path != "/" {

	}
}

func main() {
	flag.Parse()
	hub := newHub()
	go hub.run()
	//http.HandleFunc("/", serveHome)
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
