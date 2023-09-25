package main

import (
	"fmt"
	"sync"

	"golang.org/x/net/websocket"
)

const (
	ReadBuffserSize int = 1024
)

// Broadcaster that is a hub for all connections
type Broadcaster struct {
	Connections map[*websocket.Conn]bool
	events      chan []byte
	mutex       sync.RWMutex
}

// NewBroadcaster returns a new broadcaster
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		Connections: make(map[*websocket.Conn]bool),
		events:      make(chan []byte, 100),
		mutex:       sync.RWMutex{},
	}
}

// Add adds a new connection to the broadcaster
func (caster *Broadcaster) Add(conn *websocket.Conn) error {

	if conn == nil {
		return fmt.Errorf("conn cannot be nil")
	}

	caster.mutex.Lock()
	caster.Connections[conn] = true
	caster.mutex.Unlock()

	go func() {

		var n int
		var err error = nil
		var message []byte

		for err == nil {
			err = websocket.Message.Receive(conn, &message)
			fmt.Println("<<<")
			caster.events <- message[:n]
			fmt.Println(">>>")
			fmt.Println("~~~", n, err, message)
		}
	}()

	return nil
}

// Broadcast sends messages to all connections
func (caster *Broadcaster) Broadcast(msg []byte) error {
	caster.mutex.Lock()
	defer caster.mutex.Unlock()

	for conn := range caster.Connections {
		websocket.Message.Send(conn, msg)
	}

	return nil
}

// Remove removes a connection from the broadcaster
func (caster *Broadcaster) Remove(conn *websocket.Conn) error {
	delete(caster.Connections, conn)
	return nil
}

// Close closes a connection
func (caster *Broadcaster) Close(conn *websocket.Conn) error {
	return nil
}

func (caster *Broadcaster) Events() <-chan []byte {
	return nil
}
