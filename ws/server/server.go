package server

import (
	"log"
	"net/http"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
)

func liveness(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func readiness(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

// ServeOptions ...
type ServeOptions struct{}

func decodeKey(secretKey string) (*[32]byte, error) {
	if secretKey == "" {
		return nil, errors.Errorf("empty secret key")
	}
	key, err := encoding.Decode(secretKey, encoding.Hex)
	if err != nil {
		return nil, err
	}
	return keys.Bytes32(key), nil
}

// ListenAndServe starts the server.
func ListenAndServe(addr string, url string, secretKey string, opts *ServeOptions) error {
	if opts == nil {
		opts = &ServeOptions{}
	}

	sk, err := decodeKey(secretKey)
	if err != nil {
		return err
	}

	hub := NewHub(url)
	rds := NewRedis(hub, sk)
	hub.rds = rds

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
