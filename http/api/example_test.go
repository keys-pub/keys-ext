package api_test

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
)

func ExampleNewRequest() {
	key := keys.GenerateEdX25519Key()

	// Vault POST
	content := []byte(`[{"data":"dGVzdGluZzE="},{"data":"dGVzdGluZzI="}]`)
	contentHash := api.ContentHash(content)
	req, err := api.NewRequest("POST", "https://keys.pub/vault/"+key.ID().String(), bytes.NewReader(content), contentHash, time.Now(), key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("curl -H \"Authorization: %s\" -d %q %q\n", req.Header["Authorization"][0], string(content), req.URL.String())

	// Vault GET
	req, err = api.NewRequest("GET", "https://keys.pub/vault/"+key.ID().String(), nil, "", time.Now(), key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("curl -H \"Authorization: %s\" %q\n", req.Header["Authorization"][0], req.URL.String())
}
