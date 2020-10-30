package server

import (
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
func ListenAndServe(addr string, host string) error {
	hub := NewHub(host)
	rds := NewRedis(hub)
	hub.nonces = rds

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
	log.Printf("listen on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}
