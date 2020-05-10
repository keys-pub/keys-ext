package client_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/user"
	"github.com/keys-pub/keys/util"
	"github.com/keys-pub/keysd/http/client"
)

func ExampleClient_Sigchain() {
	ks := keys.NewMemStore(true)

	cl, err := client.NewClient("https://keys.pub", ks)
	if err != nil {
		log.Fatal(err)
	}

	kid := keys.ID("kex1ydecaulsg5qty2axyy770cjdvqn3ef2qa85xw87p09ydlvs5lurq53x0p3")
	resp, err := cl.Sigchain(context.TODO(), kid)
	if err != nil {
		log.Fatal(err)
	}
	sc, err := resp.Sigchain()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", sc.Spew())

	usr, err := user.FindInSigchain(sc)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", usr)

	result := user.RequestVerify(context.TODO(), util.NewHTTPRequestor(), usr, time.Now())
	if result.Status != user.StatusOK {
		log.Fatalf("User check failed: %+v", result)
	}

	fmt.Printf("%s\n", result.Status)

	// Output:
	// /sigchain/kex1ydecaulsg5qty2axyy770cjdvqn3ef2qa85xw87p09ydlvs5lurq53x0p3/1 {".sig":"Rf39uRlHXTqmOjcqUs5BTssshRdqpaFfmoYYJuA5rNVGgGd/bRG8p8ZebB8K+w9kozMpnuAoa4lko+oPHcabCQ==","data":"eyJrIjoia2V4MXlkZWNhdWxzZzVxdHkyYXh5eTc3MGNqZHZxbjNlZjJxYTg1eHc4N3AwOXlkbHZzNWx1cnE1M3gwcDMiLCJuIjoia2V5cy5wdWIiLCJzcSI6MSwic3IiOiJodHRwcyIsInUiOiJodHRwczovL2tleXMucHViL2tleXNwdWIudHh0In0=","kid":"kex1ydecaulsg5qty2axyy770cjdvqn3ef2qa85xw87p09ydlvs5lurq53x0p3","seq":1,"ts":1588276919715,"type":"user"}
	// keys.pub@https!kex1ydecaulsg5qty2axyy770cjdvqn3ef2qa85xw87p09ydlvs5lurq53x0p3-1#https://keys.pub/keyspub.txt
	// ok
}
