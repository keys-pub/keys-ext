package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/keys-pub/keys-ext/ws/server"
	"github.com/pkg/errors"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env")
	}

	// In memory nonce check
	nonces := map[string]bool{}
	opts := &server.ServeOptions{
		NonceCheck: func(ctx context.Context, nonce string) error {
			_, ok := nonces[nonce]
			if ok {
				return errors.Errorf("nonce collision")
			}
			nonces[nonce] = true
			return nil
		},
	}

	log.Fatal(server.ListenAndServe(":8080", "ws://localhost:8080/ws", opts))
}
