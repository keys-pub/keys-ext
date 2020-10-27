package main

import (
	"context"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys-ext/matter"
)

// TODO: Args for host, username, password, etc

func main() {
	matter.SetLogger(matter.NewLogger(matter.DebugLevel))

	client, err := matter.NewClient("http://localhost:8065/")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.TODO()

	if _, err = client.LoginWithPassword(ctx, "gabriel", "testpassword"); err != nil {
		log.Fatal(err)
	}

	wsClient, err := client.NewWebSocketClient()
	if err != nil {
		log.Fatal(err)
	}
	defer wsClient.Close()
	wsClient.Listen()

	for {
		select {
		case event := <-wsClient.EventChannel:
			fmt.Println("Event:")
			spew.Dump(event)
		case resp := <-wsClient.ResponseChannel:
			fmt.Println("Response:")
			spew.Dump(resp)
		case _ = <-wsClient.PingTimeoutChannel:
			log.Fatalf("timed out")
		}
	}
}
