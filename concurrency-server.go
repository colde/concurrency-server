package main

import (
	"flag"
	"log"
	"net/http"
)

var adminpw = "foobar"
var addr = flag.String("addr", ":8026", "http service address")

func main() {
	flag.Parse()
	hub := newHub()
	go hub.run()
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
