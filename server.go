package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

//go:embed static
var static embed.FS

// PromptResponse is a prompt -> response pair with some things to coordinate multiple participants.
type PromptResponseTracker struct {
	Prompt        string
	Response      string
	Secret        int
	Waiting       bool
	LastTouched   time.Time
	LastHeartbeat time.Time
	mutex         sync.Mutex
}

type PromptResponseMessage struct {
	Prompt   string `json:"prompt"`
	Response string `json:"response"`
	Waiting  bool   `json:"waiting"`
	Error    string `json:"error"`
}

// NewPromptResponseTracker returns an object to track a single prompt and its response
func NewPromptResponseTracker() *PromptResponseTracker {
	return &PromptResponseTracker{
		Prompt:   "",
		Response: "",
		Waiting:  false,
		Secret:   rand.Intn(81818181),
		mutex:    sync.Mutex{},
	}
}

// NewMux returns a new set of routes
func NewMux() (*http.ServeMux, error) {

	fs, err := fs.Sub(static, "static")

	if err != nil {
		return nil, err
	}

	tracker := NewPromptResponseTracker()

	mux := http.NewServeMux()

	mux.HandleFunc("/prompt", func(w http.ResponseWriter, req *http.Request) {

		tracker.mutex.Lock()
		defer tracker.mutex.Unlock()

		now := time.Now()

		var resp PromptResponseMessage

		switch {
		case now.Add(-10 * time.Second).Before(tracker.LastTouched):
			resp = PromptResponseMessage{
				Prompt:   "",
				Response: "",
				Waiting:  false,
				Error:    "Request happened too soon",
			}
		case tracker.Waiting:
			resp = PromptResponseMessage{
				Prompt:   "",
				Response: "",
				Waiting:  false,
				Error:    "Request occurred while another is being computed",
			}
		default:
			resp = PromptResponseMessage{
				Prompt:   "some prompt",
				Response: "",
				Waiting:  true,
				Error:    "",
			}
		}

		encoded, err := json.Marshal(resp)

		if err != nil {
			return
		}

		tracker.LastTouched = now

		w.Write(encoded)

	})

	mux.HandleFunc("/respond", func(w http.ResponseWriter, req *http.Request) {
		return
	})

	mux.Handle("/", http.FileServer(http.FS(fs)))

	return mux, nil
}

func main() {
	fmt.Println("vim-go")

	mux, _ := NewMux()

	log.Fatal(http.ListenAndServe(":8181", mux))
}
