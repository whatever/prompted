package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Event struct {
	Conn *websocket.Conn
	Data []byte
}

// BrainBleachSocialConns collects clients and broadcasts messages to clients.
type Broadcaster struct {
	Conns  map[*websocket.Conn]bool
	mutex  sync.Mutex
	events chan Event
}

// NewBrainBleachSocialConns returns a new, empty BrainBleachSocialConns obj.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		Conns:  make(map[*websocket.Conn]bool),
		mutex:  sync.Mutex{},
		events: make(chan Event),
	}
}

// Add includes a new connection and sets up a go-routine to wait for the
// connection to close.
func (conns *Broadcaster) Add(conn *websocket.Conn) {

	conns.mutex.Lock()
	conns.Conns[conn] = true
	conns.mutex.Unlock()

	go func() {
		defer func() {
			conns.Remove(conn)
		}()

		for {
			if _, m, err := conn.ReadMessage(); err == nil {
				conns.events <- Event{conn, m}
			} else {
				break
			}
		}
	}()
}

// Remove closes and removes a connection from the Hub.
func (conns *Broadcaster) Remove(conn *websocket.Conn) {
	conns.mutex.Lock()
	defer conns.mutex.Unlock()
	conn.Close()
	delete(conns.Conns, conn)
}

// Broadcast sends messages to every client.
func (conns *Broadcaster) Broadcast(message []byte) error {
	conns.mutex.Lock()
	defer conns.mutex.Unlock()
	for conn, _ := range conns.Conns {
		conn.WriteMessage(websocket.TextMessage, message)
	}
	return nil
}

// Broadcast sends messages to every client.
func (conns *Broadcaster) Send(conn *websocket.Conn, message []byte) error {
	return conn.WriteMessage(websocket.TextMessage, message)
}

// Events returns a read-only channel of valid messages sent from connections.
func (conns *Broadcaster) Events() <-chan Event {
	return conns.events
}
