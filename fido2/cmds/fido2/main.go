package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/keys-pub/keysd/fido2"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("specify fido2 library")
	}

	server, err := fido2.OpenPlugin(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	req := &fido2.DevicesRequest{}
	resp, err := server.Devices(context.TODO(), req)
	if err != nil {
		log.Fatal(err)
	}
	printResponse(resp)
}

func printResponse(i interface{}) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
