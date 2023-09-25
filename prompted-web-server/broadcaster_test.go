package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/websocket"
)

func server() *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("/echo", websocket.Handler(func(ws *websocket.Conn) {
		io.Copy(ws, ws)
	}))
	mux.Handle("/quiet", websocket.Handler(func(ws *websocket.Conn) {
	}))
	return httptest.NewServer(mux)
}

func conn(s *httptest.Server, path string) (*websocket.Conn, error) {
	return websocket.Dial(
		fmt.Sprintf("ws://%s/%s", s.Listener.Addr().String(), path),
		"",
		"ws://localhost/",
	)
}

// Test whether the broadcaster works
func TestBroadcaster(t *testing.T) {

	s := server()
	defer s.Close()

	conn, err := conn(s, "echo")

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

	fmt.Println("... <")

	if n, _ := conn.Read(bytes); string(bytes[0:n]) != "hello" {
		t.Errorf("Expected: %s Got: %s", "hello", string(bytes))
	}

	if n, _ := conn.Read(bytes); string(bytes[0:n]) != "gorgeous" {
		t.Errorf("Expected: %s Got: %s", "gorgeous", string(bytes))
	}

	fmt.Println("... >")

	c := caster.Events()

	fmt.Println(<-c)

	_ = c
}
