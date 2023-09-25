package main

import (
	"fmt"
	"sync"

	"golang.org/x/net/websocket"
)

// Broadcaster that is a hub for all connections
type Broadcaster struct {
	Connections map[*websocket.Conn]bool
	mutex       sync.RWMutex
}

// NewBroadcaster returns a new broadcaster
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		Connections: make(map[*websocket.Conn]bool),
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

	return nil
}

// Broadcast sends messages to all connections
func (caster *Broadcaster) Broadcast(msg []byte) error {
	caster.mutex.Lock()
	defer caster.mutex.Unlock()

	for conn := range caster.Connections {
		conn.Write(msg)
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

func (caster *Broadcaster) Events() chan<- []byte {
	return nil
}
