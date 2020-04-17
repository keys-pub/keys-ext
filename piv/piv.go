package main

import (
	"log"
	"strings"

	"github.com/go-piv/piv-go/piv"
	"github.com/pkg/errors"
)

func main() {
	// List all smartcards connected to the system.
	cards, err := piv.Cards()
	if err != nil {
		log.Fatal(err)
	}

	// Find a YubiKey and open the reader.
	var yk *piv.YubiKey
	for _, card := range cards {
		if strings.Contains(strings.ToLower(card), "yubikey") {
			if yk, err = piv.Open(card); err != nil {
				log.Fatal(err)
			}
			break
		}
	}
	if yk == nil {
		log.Fatal(errors.Errorf("yubikey not found"))
	}

	// Generate a private key on the YubiKey.
	key := piv.Key{
		Algorithm:   piv.AlgorithmEC256,
		PINPolicy:   piv.PINPolicyAlways,
		TouchPolicy: piv.TouchPolicyAlways,
	}
	pub, err := yk.GenerateKey(piv.DefaultManagementKey, piv.SlotAuthentication, key)
	if err != nil {
		log.Fatal(err)
	}

	auth := piv.KeyAuth{PIN: piv.DefaultPIN}
	priv, err := yk.PrivateKey(piv.SlotAuthentication, pub, auth)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("priv: %+v\n", priv)
}
