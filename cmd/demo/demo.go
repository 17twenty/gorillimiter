package main

import (
	"log"
	"net/http"
	"time"

	"github.com/17twenty/gorillimiter"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Wrap the servemux with the limit middleware.
	log.Println("Listening on :5000...")
	http.ListenAndServe(":5000", gorillimiter.Limiter(mux, 10, time.Second))
}
