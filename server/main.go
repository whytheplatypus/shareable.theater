package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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

	defaultMaxOccupancy    = "150"
	defaultMaxTheaterCount = "10000"
)

func fromEnv(key string, or string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return or
	}
	return val
}

func must(val interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return val
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var cinema = map[string]*Theater{}

var addr = flag.String("addr", ":8080", "http service address")
var maxOccupancy = flag.Int("occupancy", must(strconv.Atoi(fromEnv("MAX_OCCUPANCY", defaultMaxOccupancy))).(int), "the max number of connected audience members a theater can have")
var maxTheaterCount = flag.Int("theater-count", must(strconv.Atoi(fromEnv("MAX_THEATER_COUNT", defaultMaxTheaterCount))).(int), "the max number of concurrent theaters this server will support")

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
		if ok && cinema[theater].Projectionist != nil {
			buf := &bytes.Buffer{}
			tmpl.Execute(buf, "")
			inUsePage.Execute(w, buf.String())
			w.WriteHeader(http.StatusConflict)
			return
		}
		if len(cinema) > *maxTheaterCount {
			http.ServeFile(w, r, "../static/full-building.html")
			return
		}
		http.ServeFile(w, r, "../static/projectionist/index.html")
	})
	r.HandleFunc("/projectionist/{theater}/signal", func(w http.ResponseWriter, r *http.Request) {
		if len(cinema) > *maxTheaterCount {
			http.Error(w, "The whole building is full", http.StatusServiceUnavailable)
			return
		}
		vars := mux.Vars(r)
		theater := vars["theater"]
		_, ok := cinema[theater]
		if ok && cinema[theater].Projectionist != nil {
			http.Error(w, "That theater is occupied", http.StatusConflict)
			return
		}

		if cinema[theater] == nil {
			// creation
			// todo make this more prominant
			cinema[theater] = &Theater{
				Audience: map[chan []byte]bool{},
			}
			// starts the "setup" wait
			cinema[theater].wg.Add(1)
			// finishes the "setup" wait
			defer cinema[theater].wg.Done()
			go func(theater string) {
				cinema[theater].wg.Wait()
				delete(cinema, theater)
				log.Println("Removing", theater)
			}(theater)
		}

		cinema[theater].Projectionist = make(chan []byte, 256)
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
		if len(cinema[theater].Audience) > *maxOccupancy {
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
