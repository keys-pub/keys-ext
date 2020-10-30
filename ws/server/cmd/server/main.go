package main

import (
	"log"

	"github.com/keys-pub/keys-ext/ws/server"
)

func main() {
	log.Fatal(server.ListenAndServe(":8080", "localhost"))
}
