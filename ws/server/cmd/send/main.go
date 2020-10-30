package main

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/ws/api"
	"github.com/keys-pub/keys-ext/ws/server"
)

func main() {
	redisPool := server.NewRedisPool()
	redisConn := redisPool.Get()
	defer redisConn.Close()

	send := func(msg *api.Message) error {
		b, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		if _, err := redisConn.Do("PUBLISH", "message", b); err != nil {
			return err
		}
		return nil
	}

	for i := 0; i < 100; i = i + 10 {
		key := keys.NewEdX25519KeyFromSeed(testSeed(byte(i)))
		if err := send(&api.Message{KID: key.ID()}); err != nil {
			log.Fatal(err)
		}
	}
}

func testSeed(b byte) *[32]byte {
	return keys.Bytes32(bytes.Repeat([]byte{b}, 32))
}
