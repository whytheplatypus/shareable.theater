package main

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Room struct {
	Host    chan []byte
	Clients map[chan []byte]bool
}

const DEMO_ROOM = "Hello Room"

var rooms = map[string]*Room{}

var upgrader = websocket.Upgrader{}

func serveViewer(w http.ResponseWriter, r *http.Request) {
	room, ok := rooms[DEMO_ROOM]
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

func serveHost(w http.ResponseWriter, r *http.Request) {
	_, ok := rooms[DEMO_ROOM]
	if ok {
		log.Println("room is taken")
		return
	}

	room := &Room{
		Host:    make(chan []byte, 256),
		Clients: map[chan []byte]bool{},
	}
	rooms[DEMO_ROOM] = room

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
			delete(rooms, DEMO_ROOM)
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
	http.Handle("/booth/", http.StripPrefix("/booth/", http.FileServer(http.Dir("../booth"))))
	http.Handle("/audience/", http.StripPrefix("/audience/", http.FileServer(http.Dir("../audience"))))
	http.Handle("/shared/", http.StripPrefix("/shared/", http.FileServer(http.Dir("../shared"))))
	http.HandleFunc("/host", func(w http.ResponseWriter, r *http.Request) {
		serveHost(w, r)
	})
	http.HandleFunc("/viewer", func(w http.ResponseWriter, r *http.Request) {
		serveViewer(w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
