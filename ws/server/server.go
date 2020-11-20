package server

import (
	"log"
	"net/http"
	"time"

	"github.com/keys-pub/keys-ext/ws/api"
)

func liveness(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func readiness(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// ServeOptions ...
type ServeOptions struct {
	// NonceCheck to override default nonce check.
	NonceCheck api.NonceCheck
}

// ListenAndServe starts the server.
func ListenAndServe(addr string, url string, opts *ServeOptions) error {
	if opts == nil {
		opts = &ServeOptions{}
	}

	hub := NewHub(url)
	rds := NewRedis(hub)
	hub.rds = rds

	if opts.NonceCheck != nil {
		hub.nonceCheck = opts.NonceCheck
	}

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
