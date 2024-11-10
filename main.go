package main

import (
	"time"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	activeUsersLock sync.Mutex
	activeUsers = make(map[string]bool)

	messagesLock sync.Mutex
	messages = []string{}
)

func main() {
	addr := fmt.Sprintf(
		"%s:%s",
		os.Getenv("CHAT_ADDR"),
		os.Getenv("CHAT_PORT"),
	)

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

		activeUsersLock.Lock()
		activeUsers[name] = true
		activeUsersLock.Unlock()
	})

	http.HandleFunc("POST /message", func(w http.ResponseWriter, r *http.Request) {
		messagesLock.Lock()
		messages = append(messages, r.PostFormValue("message"))
		messagesLock.Unlock()
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

		i := 1
		for {
			fmt.Fprintf(w, "data: <div>Message %d at %s</div>\n\n", i, time.Now().Format(time.RFC3339))
			flusher.Flush()
			time.Sleep(1 * time.Second)
			i += 1
		}
	})

	fmt.Printf("Running server on: %s\n", addr)
	http.ListenAndServe(addr, nil)
}
