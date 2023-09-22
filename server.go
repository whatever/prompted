package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"
)

//go:embed static
var static embed.FS

// PromptResponse is a prompt -> response pair with some things to coordinate multiple participants.
type PromptResponseTracker struct {
	Prompt        string     `json:"prompt"`
	Response      string     `json:"response"`
	Secret        string     `json:"-"`
	Waiting       bool       `json:"waiting"`
	LastTouched   time.Time  `json:"-"`
	LastHeartbeat time.Time  `json:"-"`
	mutex         sync.Mutex `json:"-"`
}

// PromptResponseMessage is a response type which is probably redundant with the above
type PromptResponseMessage struct {
	Prompt   string `json:"prompt"`
	Response string `json:"response"`
	Waiting  bool   `json:"waiting"`
	Error    string `json:"error,omitempty"`
}

// NewPromptResponseTracker returns an object to track a single prompt and its response
func NewPromptResponseTracker() *PromptResponseTracker {
	return &PromptResponseTracker{
		Prompt:   "",
		Response: "",
		Waiting:  false,
		// Secret:   fmt.Sprintf("%d", rand.Intn(81818181)),
		Secret: "8181",
		mutex:  sync.Mutex{},
	}
}

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
		case now.Add(-30 * time.Second).Before(tracker.LastTouched):
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

		case !req.Form.Has("prompt"):
			resp = PromptResponseMessage{
				Prompt:   "",
				Response: "",
				Waiting:  false,
				Error:    "Request is missing prompt in post form data field",
			}

		default:
			resp = PromptResponseMessage{
				Prompt:   req.Form.Get("prompt"),
				Response: "",
				Waiting:  true,
				Error:    "",
			}
			tracker.LastTouched = now
			tracker.Prompt = req.Form.Get("prompt")
		}

		encoded, err := json.Marshal(resp)

		if err != nil {
			return
		}

		w.Write(encoded)

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
			errmsg = "request is not sending the correct secret"

		default:
			tracker.Response = req.Form.Get("response")
			tracker.LastTouched = now
		}

		resp := PromptResponseMessage{
			Prompt:   tracker.Prompt,
			Response: tracker.Response,
			Waiting:  tracker.Waiting,
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
			Waiting:  tracker.Waiting,
			Error:    "",
		})
	})

	mux.Handle("/", http.FileServer(http.FS(fs)))

	return mux, nil
}

func main() {
	mux, _ := NewMux()

	log.Println("Listening on port 8181")

	log.Fatal(http.ListenAndServe(":8181", mux))
}
