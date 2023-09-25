package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type TrackerState string

const (
	TrackerStateWaiting TrackerState = "waiting"
	TrackerStateWorking TrackerState = "working"
	TrackerStateReady   TrackerState = "ready"
)

//go:embed static
var static embed.FS

// PromptResponse is a prompt -> response pair with some things to coordinate multiple participants.
type PromptResponseTracker struct {
	Prompt        string     `json:"prompt"`
	Response      string     `json:"response"`
	Secret        string     `json:"-"`
	State         string     `json:"state"`
	LastTouched   time.Time  `json:"-"`
	LastHeartbeat time.Time  `json:"-"`
	mutex         sync.Mutex `json:"-"`
}

// PromptResponseMessage is a response type which is probably redundant with the above
type PromptResponseMessage struct {
	Prompt   string `json:"prompt"`
	Response string `json:"response"`
	State    string `json:"state"`
	Error    string `json:"error,omitempty"`
}

// NewPromptResponseTracker returns an object to track a single prompt and its response
func NewPromptResponseTracker() *PromptResponseTracker {
	return &PromptResponseTracker{
		Prompt:   "",
		Response: "",
		State:    "ready",
		Secret:   "8181", // fmt.Sprintf("%d", rand.Intn(81818181)),
		mutex:    sync.Mutex{},
	}
}

// StatusMessage returns a message with the current state/status of the tracker
func (tracker *PromptResponseTracker) StatusMessage() PromptResponseMessage {
	return PromptResponseMessage{
		Prompt:   tracker.Prompt,
		Response: tracker.Response,
		State:    tracker.State,
		Error:    "",
	}
}

var port = flag.Int("port", 8182, "set port to listen on")

func main() {
	mux, err := NewMux()

	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()

	log.Printf("Listening on port %d", *port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), mux))
}
