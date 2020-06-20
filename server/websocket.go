package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type Theater struct {
	Projectionist chan []byte
	Audience      map[chan []byte]bool
}

func (c *Theater) audienceWebsocket(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := make(chan []byte, 256)
	c.Audience[client] = true
	messages := make(chan []byte, 256)

	// read
	go read(messages, conn)
	go func(messages <-chan []byte) {
		defer delete(c.Audience, client)
		defer recoverln("Recovered from audience websocket read handler")
		for {
			message, ok := <-messages
			if !ok {
				return
			}
			log.Println("clients got message", string(message))
			log.Println("clients sent message to host", string(message))
			c.Projectionist <- message
		}
	}(messages)

	go write(client, conn)
}

func (c *Theater) projectionistWebsocket(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// from https://github.com/gorilla/websocket/blob/master/examples/chat/client.go
	messages := make(chan []byte, 256)
	// read
	go read(messages, conn)
	go func(messages <-chan []byte) {
		defer recoverln("Recovered from audience websocket read handler")
		for {
			message, ok := <-messages
			if !ok {
				return
			}
			log.Println("host got message", string(message))
			for client, _ := range c.Audience {
				go func(client chan []byte, message []byte) {
					defer recoverln("Recovered during write to client")
					client <- message
				}(client, message)
			}
		}
	}(messages)
	// write
	go write(c.Projectionist, conn)
}

func recoverln(msg string) {
	if r := recover(); r != nil {
		log.Println(msg, r)
	}
}

func write(messages <-chan []byte, conn *websocket.Conn) {
	defer recoverln("Recovered from audience websocket write handler")

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()
	for {
		select {
		case message, ok := <-messages:
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
			n := len(messages)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-messages)
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
}

func read(messages chan<- []byte, conn *websocket.Conn) {
	defer conn.Close()
	defer close(messages)
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
		messages <- message
	}
}
