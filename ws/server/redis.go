package server

import (
	"encoding/json"
	"log"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/keys-pub/keys-ext/ws/api"
)

// NewRedisPool from env.
func NewRedisPool() *redis.Pool {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", redisAddr)
			if redisPassword == "" {
				return conn, err
			}
			if err != nil {
				return nil, err
			}
			if _, err := conn.Do("AUTH", redisPassword); err != nil {
				conn.Close()
				return nil, err
			}
			return conn, nil
		},
		// TODO: Tune other settings, like IdleTimeout, MaxActive, MaxIdle, TestOnBorrow.
	}
}

// Redis ...
type Redis struct {
	redisPool *redis.Pool
	hub       *Hub
}

// NewRedis ...
func NewRedis(hub *Hub) *Redis {
	redisPool := NewRedisPool()
	return &Redis{
		redisPool: redisPool,
		hub:       hub,
	}
}

// Subscribe ...
func (r *Redis) Subscribe() error {
	redisConn := r.redisPool.Get()
	defer redisConn.Close()

	log.Printf("subscribe\n")
	psc := redis.PubSubConn{Conn: redisConn}
	psc.Subscribe("message")
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			log.Printf("channel %s: message: %s\n", v.Channel, v.Data)
			var msg api.Message
			if err := json.Unmarshal(v.Data, &msg); err != nil {
				log.Printf("error receiving message: %v", v)
				break
			}
			r.hub.broadcast <- &msg
		case redis.Subscription:
			log.Printf("subscription %s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			return v
		}
	}
}
