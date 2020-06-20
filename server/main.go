package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "net/http/pprof"

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

var cinema = map[string]*Theater{}

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/projectionist/", func(w http.ResponseWriter, r *http.Request) {
		buf := &bytes.Buffer{}
		tmpl.Execute(buf, "")
		theater := fmt.Sprintf("/projectionist/%s", buf.String())
		http.Redirect(w, r, theater, 302)
	})
	r.HandleFunc("/projectionist/{theater}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		theater := vars["theater"]
		_, ok := cinema[theater]
		if ok {
			log.Println("room is already taken")
			buf := &bytes.Buffer{}
			tmpl.Execute(buf, "")
			inUsePage.Execute(w, buf.String())
			w.WriteHeader(http.StatusConflict)
			return
		}
		if len(cinema) > THEATER_LIMIT {
			http.Error(w, "The whole building is full", http.StatusServiceUnavailable)
			return
		}
		http.ServeFile(w, r, "../static/projectionist/index.html")
	})
	r.HandleFunc("/projectionist/{theater}/signal", func(w http.ResponseWriter, r *http.Request) {
		if len(cinema) > THEATER_LIMIT {
			http.Error(w, "The whole building is full", http.StatusServiceUnavailable)
			return
		}
		vars := mux.Vars(r)
		theater := vars["theater"]
		_, ok := cinema[theater]
		if ok {
			http.Error(w, "That theater is occupied", http.StatusConflict)
			return
		}

		cinema[theater] = &Theater{
			Projectionist: make(chan []byte, 256),
			Audience:      map[chan []byte]bool{},
		}
		cinema[theater].projectionistWebsocket(w, r)
	})

	r.HandleFunc("/audience/{theater}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		theater := vars["theater"]
		_, ok := cinema[theater]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			http.ServeFile(w, r, "../static/empty.html")
			return
		}
		if len(cinema[theater].Audience) > THEATER_CAPACITY {
			http.ServeFile(w, r, "../static/full.html")
			return
		}
		http.ServeFile(w, r, "../static/audience/index.html")
	})

	r.HandleFunc("/audience/{theater}/signal", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		theater := vars["theater"]
		room, ok := cinema[theater]
		if !ok {
			http.NotFound(w, r)
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
