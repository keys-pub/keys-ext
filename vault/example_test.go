package vault_test

import (
	"fmt"
	"log"
	"time"

	"github.com/keys-pub/keys-ext/vault"
)

func ExampleNew() {
	// New vault.
	// Use vault.NewDB for a persistent store.
	vlt := vault.New(vault.NewMem())
	if err := vlt.Open(); err != nil {
		log.Fatal(err)
	}
	defer vlt.Close()

	// Setup auth.
	if err := vlt.UnlockWithPassword("mypassword", true); err != nil {
		log.Fatal(err)
	}

	// Save item.
	// Item IDs are NOT encrypted locally and are provided for fast lookups.
	item := vault.NewItem("id1", []byte("mysecret"), "", time.Now())
	if err := vlt.Set(item); err != nil {
		log.Fatal(err)
	}

	// Get item.
	out, err := vlt.Get("id1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("secret: %s\n", string(out.Data))

	// List items.
	items, err := vlt.Items()
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range items {
		fmt.Printf("%s: %v\n", item.ID, string(item.Data))
	}

	// Output:
	// secret: mysecret
	// id1: mysecret
}
