package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/keys-pub/keys-ext/ws/server"
)

func main() {
	flag.Parse()

	hub := server.NewHub()
	rds := server.NewRedis(hub)

	go func() {
		for {
			if err := rds.Subscribe(); err != nil {
				log.Printf("error in subscribe: %v", err)
				time.Sleep(time.Second * 2)
			}
		}
	}()

	go hub.Run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.Serve(hub, w, r)
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("error in ListenAndServe: %v", err)
	}
}
