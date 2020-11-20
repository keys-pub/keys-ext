package main

import (
	"bytes"
	"log"

	"github.com/joho/godotenv"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/ws/api"
	"github.com/keys-pub/keys-ext/ws/server"
	"github.com/vmihailenco/msgpack/v4"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env")
	}

	redisPool := server.NewRedisPool()
	redisConn := redisPool.Get()
	defer redisConn.Close()

	send := func(event *api.PubEvent) error {
		b, err := msgpack.Marshal(event)
		if err != nil {
			return err
		}
		if _, err := redisConn.Do("PUBLISH", api.EventPubSub, b); err != nil {
			return err
		}
		return nil
	}

	channel := keys.NewEdX25519KeyFromSeed(testSeed(0xef))

	ids := []keys.ID{}
	for i := 0; i < 20; i++ {
		user := keys.NewEdX25519KeyFromSeed(testSeed(byte(i)))
		ids = append(ids, user.ID())
	}

	if err := send(&api.PubEvent{Channel: channel.ID(), Users: ids, Index: 1}); err != nil {
		log.Fatal(err)
	}
}

func testSeed(b byte) *[32]byte {
	return keys.Bytes32(bytes.Repeat([]byte{b}, 32))
}
