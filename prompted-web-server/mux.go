package main

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"time"
)

// NewMux returns a new set of routes
func NewMux() (*http.ServeMux, error) {

	fs, err := fs.Sub(static, "static")

	if err != nil {
		return nil, err
	}

	tracker := NewPromptResponseTracker()

	log.Printf("Initialized with secret %v", tracker.Secret)

	mux := http.NewServeMux()

	mux.HandleFunc("/prompt", func(w http.ResponseWriter, req *http.Request) {

		tracker.mutex.Lock()
		defer tracker.mutex.Unlock()

		now := time.Now()

		var resp PromptResponseMessage

		req.ParseForm()

		switch {
		case now.Add(-3 * time.Second).Before(tracker.LastTouched):
			resp = PromptResponseMessage{
				Prompt:   "",
				Response: "",
				State:    tracker.State,
				Error:    "Request happened too soon",
			}

		case tracker.State == "working":
			resp = PromptResponseMessage{
				Prompt:   "",
				Response: "",
				State:    tracker.State,
				Error:    "Request occurred while another is being computed",
			}

		case !req.Form.Has("prompt"):
			resp = PromptResponseMessage{
				Prompt:   "",
				Response: "",
				State:    tracker.State,
				Error:    "Request is missing prompt in post form data field",
			}

		default:
			resp = PromptResponseMessage{
				Prompt:   req.Form.Get("prompt"),
				Response: "",
				State:    tracker.State,
				Error:    "",
			}
			tracker.LastTouched = now
			tracker.State = "waiting"
			tracker.Prompt = req.Form.Get("prompt")
			tracker.Response = ""
		}

		log.Println("!!!!")

		encoded, err := json.Marshal(resp)

		if err != nil {
			return
		}

		w.Write(encoded)

	})

	mux.HandleFunc("/heartbeat", func(w http.ResponseWriter, req *http.Request) {
		tracker.mutex.Lock()
		defer tracker.mutex.Unlock()

		req.ParseForm()

		errmsg := ""
		now := time.Now()

		switch {
		case req.Form.Get("secret") != tracker.Secret:
			errmsg = "provided secret is incorrect"
		case !req.Form.Has("state"):
			errmsg = "request is missing state parameter"
		default:
			tracker.State = req.Form.Get("state")
			tracker.LastHeartbeat = now
		}

		if errmsg != "" {
			log.Printf("error: %v", errmsg)
		}

		json.NewEncoder(w).Encode(PromptResponseMessage{
			Prompt:   tracker.Prompt,
			Response: tracker.Response,
			State:    tracker.State,
			Error:    errmsg,
		})

	})

	mux.HandleFunc("/respond", func(w http.ResponseWriter, req *http.Request) {

		tracker.mutex.Lock()
		defer tracker.mutex.Unlock()

		now := time.Now()

		req.ParseForm()

		errmsg := ""

		switch {
		case !req.Form.Has("prompt"):
			errmsg = "request is missing prompt parameter"

		case !req.Form.Has("response"):
			errmsg = "request is missing response parameter"

		case !req.Form.Has("secret"):
			errmsg = "request is missing secret parameter"

		case req.Form.Get("prompt") != tracker.Prompt:
			log.Printf("%v != %v", req.Form.Get("prompt"), tracker.Prompt)
			errmsg = "request is responding to the wrong prompt"

		case req.Form.Get("secret") != tracker.Secret:
			log.Printf("secret is incorrect")
			errmsg = "request is not sending the correct secret"

		default:
			log.Println("!???!")
			tracker.Response = req.Form.Get("response")
			tracker.LastTouched = now
		}

		resp := PromptResponseMessage{
			Prompt:   tracker.Prompt,
			Response: tracker.Response,
			State:    tracker.State,
			Error:    errmsg,
		}

		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, req *http.Request) {
		tracker.mutex.Lock()
		defer tracker.mutex.Unlock()
		json.NewEncoder(w).Encode(PromptResponseMessage{
			Prompt:   tracker.Prompt,
			Response: tracker.Response,
			State:    tracker.State,
			Error:    "",
		})
	})

	mux.Handle("/", http.FileServer(http.FS(fs)))

	return mux, nil
}
