package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/ws/api"
	"github.com/keys-pub/keys-ext/ws/server"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
)

func decodeKey(secretKey string) (*[32]byte, error) {
	if secretKey == "" {
		return nil, errors.Errorf("empty secret key")
	}
	key, err := encoding.Decode(secretKey, encoding.Hex)
	if err != nil {
		return nil, err
	}
	return keys.Bytes32(key), nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env")
	}

	secretKey, err := decodeKey(os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	redisPool := server.NewRedisPool()
	redisConn := redisPool.Get()
	defer redisConn.Close()

	send := func(event *api.Event) error {
		b, err := api.Encrypt(event, secretKey)
		if err != nil {
			return err
		}
		if _, err := redisConn.Do("PUBLISH", api.EventPubSub, b); err != nil {
			return err
		}
		return nil
	}

	for i := 0; i < 20; i += 2 {
		channel := keys.NewEdX25519KeyFromSeed(testSeed(byte(i)))
		token := fmt.Sprintf("testtoken%d", i)
		if err := send(&api.Event{KID: channel.ID(), Index: 1, Token: token}); err != nil {
			log.Fatal(err)
		}
	}
}

func testSeed(b byte) *[32]byte {
	return keys.Bytes32(bytes.Repeat([]byte{b}, 32))
}
