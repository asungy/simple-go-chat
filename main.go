package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	addr := fmt.Sprintf(
		"%s:%s",
		os.Getenv("CHAT_ADDR"),
		os.Getenv("CHAT_PORT"),
	)

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "html/welcome.html")
	})

	http.HandleFunc("POST /chat", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "html/chat.html")
	})

	fmt.Printf("Running server on: %s\n", addr)
	http.ListenAndServe(addr, nil)
}
