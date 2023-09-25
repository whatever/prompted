package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/websocket"
)

func TestNewBroadcaster(t *testing.T) {

	caster := NewBroadcaster()

	if caster == nil {
		t.Errorf("NewBroadcaster() returned nil")
	}

	if err := caster.Add(nil); err == nil {
		t.Errorf("Add(nil) should return an error")
	}

	conn := websocket.Conn{}

	if err := caster.Add(&conn); err != nil {
		t.Errorf("Add(&conn) returned an error")
	}
}

func server() *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("/echo", websocket.Handler(func(ws *websocket.Conn) {
		io.Copy(ws, ws)
	}))
	return httptest.NewServer(mux)
}

func conn(s *httptest.Server) (*websocket.Conn, error) {
	return websocket.Dial(
		fmt.Sprintf("ws://%s/echo", s.Listener.Addr().String()),
		"",
		"ws://localhost/",
	)
}

func TestConn(t *testing.T) {

	s := server()
	defer s.Close()

	conn, err := conn(s)

	if err != nil {
		t.Errorf("conn(s) returned an error")
	}

	caster := NewBroadcaster()

	if _ = caster.Add(conn); len(caster.Connections) != 1 {
		t.Errorf("Add(conn) did not add the connection")
	}

	caster.Broadcast([]byte("hello"))
	caster.Broadcast([]byte("gorgeous"))

	bytes := make([]byte, 1024)

	if n, _ := conn.Read(bytes); string(bytes[0:n]) != "hello" {
		t.Errorf("Expected: %s Got: %s", "hello", string(bytes))
	}

	if n, _ := conn.Read(bytes); string(bytes[0:n]) != "gorgeous" {
		t.Errorf("Expected: %s Got: %s", "gorgeous", string(bytes))
	}

	c := caster.Events()

	_ = c
}
