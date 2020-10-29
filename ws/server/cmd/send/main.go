package main

import (
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

	aid := keys.ID("kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077")
	if err := send(&api.Message{KID: aid}); err != nil {
		log.Fatal(err)
	}

	bid := keys.ID("kex1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2quf6zcg")
	if err := send(&api.Message{KID: bid}); err != nil {
		log.Fatal(err)
	}

}
