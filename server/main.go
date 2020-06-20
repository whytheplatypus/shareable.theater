package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	THEATER_CAPACITY = 50
	THEATER_LIMIT    = 1000
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var rooms = map[string]*Room{}

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/booth/", func(w http.ResponseWriter, r *http.Request) {
		buf := &bytes.Buffer{}
		tmpl.Execute(buf, "")
		theater := fmt.Sprintf("/booth/%s", buf.String())
		http.Redirect(w, r, theater, 302)
	})
	r.HandleFunc("/booth/{theater}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		theater := vars["theater"]
		_, ok := rooms[theater]
		if ok {
			log.Println("room is already taken")
			buf := &bytes.Buffer{}
			tmpl.Execute(buf, "")
			inUsePage.Execute(w, buf.String())
			w.WriteHeader(http.StatusConflict)
			return
		}
		if len(rooms) > THEATER_LIMIT {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		http.ServeFile(w, r, "../static/booth/index.html")
	})
	r.HandleFunc("/booth/{theater}/signal", func(w http.ResponseWriter, r *http.Request) {
		if len(rooms) > THEATER_LIMIT {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		vars := mux.Vars(r)
		theater := vars["theater"]
		_, ok := rooms[theater]
		if ok {
			log.Println("room is taken")
			return
		}

		rooms[theater] = &Room{
			Host:    make(chan []byte, 256),
			Clients: map[chan []byte]bool{},
		}
		rooms[theater].projectionistWebsocket(w, r)
	})

	r.HandleFunc("/audience/{theater}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		theater := vars["theater"]
		_, ok := rooms[theater]
		if !ok {
			http.ServeFile(w, r, "../static/empty.html")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if len(rooms[theater].Clients) > THEATER_CAPACITY {
			http.ServeFile(w, r, "../static/full.html")
			return
		}
		http.ServeFile(w, r, "../static/audience/index.html")
	})

	r.HandleFunc("/audience/{theater}/signal", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		theater := vars["theater"]
		room, ok := rooms[theater]
		if !ok {
			// 404
			return
		}
		room.audienceWebsocket(w, r)
	})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("../static"))))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../static/index.html")
	})
	http.Handle("/", r)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
