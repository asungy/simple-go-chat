package main

import (
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

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		name, err := r.Cookie("Name")
		if err != nil {
			tmpl.ExecuteTemplate(w, "index", map[string]any{
				"content": "welcome",
			})
		} else {
			tmpl.ExecuteTemplate(w, "chat", map[string]any{
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

	fmt.Printf("Running server on: %s\n", addr)
	http.ListenAndServe(addr, nil)
}
