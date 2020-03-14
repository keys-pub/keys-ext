package main

import (
	"log"

	"github.com/keys-pub/keysd/wormhole"
)

func main() {
	client := wormhole.NewClient()

	client.SetPublicAddressLn(func(addr string) {

	})

	client.SetPeer()

	if err := client.Listen(); err != nil {
		log.Fatal(err)
	}

}
