package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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

type Room struct {
	Host    chan []byte
	Clients map[chan []byte]bool
}

var rooms = map[string]*Room{}

var upgrader = websocket.Upgrader{}

func serveViewer(w http.ResponseWriter, r *http.Request, theater string) {
	room, ok := rooms[theater]
	if !ok {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := make(chan []byte, 256)
	room.Clients[client] = true

	// read
	go func(room *Room, conn *websocket.Conn) {
		defer conn.Close()
		conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			}
			message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
			log.Println("clients got message", string(message))
			log.Println("clients sent message to host", string(message))
			room.Host <- message
		}
	}(room, conn)

	go func(client chan []byte, conn *websocket.Conn) {
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			conn.Close()
			delete(room.Clients, client)
			close(client)
		}()
		for {
			select {
			case message, ok := <-client:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					// The hub closed the channel.
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				w, err := conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				w.Write(message)

				// Add queued chat messages to the current websocket message.
				n := len(client)
				for i := 0; i < n; i++ {
					w.Write(newline)
					w.Write(<-client)
				}

				if err := w.Close(); err != nil {
					return
				}
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}(client, conn)
}

func serveHost(w http.ResponseWriter, r *http.Request, theater string) {
	_, ok := rooms[theater]
	if ok {
		log.Println("room is taken")
		return
	}

	room := &Room{
		Host:    make(chan []byte, 256),
		Clients: map[chan []byte]bool{},
	}
	rooms[theater] = room

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// from https://github.com/gorilla/websocket/blob/master/examples/chat/client.go
	// read
	go func(room *Room, conn *websocket.Conn) {
		defer conn.Close()
		conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			}
			message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
			log.Println("host got message", string(message))
			for client, _ := range room.Clients {
				go func(client chan []byte, message []byte) {
					client <- message
				}(client, message)
			}
		}
	}(room, conn)
	// write
	go func(room *Room, conn *websocket.Conn) {
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			conn.Close()
			log.Println("closing host")
			close(room.Host)
			delete(rooms, theater)
		}()
		for {
			select {
			case message, ok := <-room.Host:
				log.Println("message sent to host")
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					log.Println("channel closed")
					// The hub closed the channel.
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				w, err := conn.NextWriter(websocket.TextMessage)
				if err != nil {
					log.Println(err)
					return
				}
				w.Write(message)

				// Add queued chat messages to the current websocket message.
				n := len(room.Host)
				for i := 0; i < n; i++ {
					w.Write(newline)
					w.Write(<-room.Host)
				}

				if err := w.Close(); err != nil {
					log.Println(err)
					return
				}
			case <-ticker.C:
				log.Println("host ticker fired")
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Println(err)
					return
				}
			}
		}
	}(room, conn)
}

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
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("../static"))))

	r.HandleFunc("/booth/{theater}/signal", func(w http.ResponseWriter, r *http.Request) {
		if len(rooms) > THEATER_LIMIT {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		vars := mux.Vars(r)
		theater := vars["theater"]
		serveHost(w, r, theater)
	})

	r.HandleFunc("/audience/{theater}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		theater := vars["theater"]
		_, ok := rooms[theater]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			http.ServeFile(w, r, "../static/empty.html")
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
		serveViewer(w, r, theater)
	})
	http.Handle("/", r)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
