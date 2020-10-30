package server

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func liveness(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func readiness(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// ListenAndServe starts the server.
func ListenAndServe(port int) {
	hub := NewHub()
	rds := NewRedis(hub)

	go func() {
		for {
			if err := rds.Subscribe(); err != nil {
				log.Printf("error in subscribe: %v", err)
				time.Sleep(time.Second * 2)
			}
		}
	}()

	go hub.Run()
	http.HandleFunc("/liveness_check", liveness)
	http.HandleFunc("/readiness_check", readiness)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		Serve(hub, w, r)
	})
	addr := fmt.Sprintf(":%d", port)
	log.Printf("listen on %s\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("error in ListenAndServe: %v", err)
	}
}
