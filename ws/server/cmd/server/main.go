package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/keys-pub/keys-ext/ws/server"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env")
	}

	sk := os.Getenv("SECRET_KEY")
	log.Fatal(server.ListenAndServe(":8080", "ws://localhost:8080/ws", sk, nil))
}
