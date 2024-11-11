package main

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var addr string
var est *time.Location
func init() {
	var err error
	est, err = time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}

	addr = fmt.Sprintf(
		"%s:%s",
		os.Getenv("CHAT_ADDR"),
		os.Getenv("CHAT_PORT"),
	)

}

type Event interface {
	SseString() string
}

type Message struct {
	name string
	message string
	timestamp time.Time
}

func (m Message) SseString() string {
	return fmt.Sprintf(
		"data: <div><b>%s</b> (%s): %s</div>\n\n",
		html.EscapeString(m.name),
		m.timestamp.In(est).Format("15:04:05"),
		html.EscapeString(m.message),
	)
}

type Join struct {
	name string
	timestamp time.Time
}

func (j Join) SseString() string {
	return fmt.Sprintf(
		"data: <div style=\"color: green;\">%s has joined the chat!</div>\n\n",
		html.EscapeString(j.name),
	)
}

type Broadcaster struct {
	chanList []chan<- Event
	l sync.Mutex
}

func (b *Broadcaster) AddConn(ch chan<- Event) {
	b.l.Lock()
	defer b.l.Unlock()

	b.chanList = append(b.chanList, ch)
}

func (b *Broadcaster) BroadcastEvent(event Event) {
	b.l.Lock()
	defer b.l.Unlock()

	for _, ch := range b.chanList {
		ch <- event
	}
}

var broadcaster Broadcaster

func main() {
	tmpl, err := template.ParseFiles(
		"templates/chat.html",
		"templates/index.html",
		"templates/welcome.html",
	)
	if err != nil {
		log.Fatalf("Error parsing template files: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name, err := r.Cookie("Name")
		if err != nil {
			tmpl.ExecuteTemplate(w, "index", map[string]any{
				"content": "welcome",
			})
		} else {
			tmpl.ExecuteTemplate(w, "index", map[string]any{
				"content": "chat",
				"name": name.Value,
			})
		}
	})

	http.HandleFunc("POST /chat", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		http.SetCookie(w, &http.Cookie{
			Name:        "Name",
			Value:       name,
			Path:        "/",
		})
		tmpl.ExecuteTemplate(w, "chat", map[string]any{
			"name": name,
		})

		broadcaster.BroadcastEvent(Join{
			name:      name,
			timestamp: time.Now(),
		})
	})

	http.HandleFunc("POST /message", func(w http.ResponseWriter, r *http.Request) {
		name := func() string {
			cookie, _ := r.Cookie("Name")
			return cookie.Value
		}()
		message := r.PostFormValue("message")
		messageEvent := Message{
			name:      name,
			message:   message,
			timestamp: time.Now(),
		}
		broadcaster.BroadcastEvent(messageEvent)
	})

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		ch := make(chan Event)
		broadcaster.AddConn(ch)

		for {
			select {
			case event := <-ch:
				w.Write([]byte(event.SseString()))
				flusher.Flush()
			}
		}
	})

	fmt.Printf("Running server on: %s\n", addr)
	http.ListenAndServe(addr, nil)
}
