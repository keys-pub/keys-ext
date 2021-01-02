package server

import (
	"context"
	"log"
	"os"
	"time"

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
				_ = conn.Close()
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
	secretKey *[32]byte
}

// NewRedis ...
func NewRedis(hub *Hub, secretKey *[32]byte) *Redis {
	redisPool := NewRedisPool()
	return &Redis{
		redisPool: redisPool,
		hub:       hub,
		secretKey: secretKey,
	}
}

// Subscribe ...
func (r *Redis) Subscribe() error {
	redisConn := r.redisPool.Get()
	defer redisConn.Close()

	log.Printf("subscribe\n")
	psc := redis.PubSubConn{Conn: redisConn}
	if err := psc.Subscribe(api.EventPubSub); err != nil {
		return err
	}
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			log.Printf("channel %s (%d)\n", v.Channel, len(v.Data))
			var event api.Event
			if err := api.Decrypt(v.Data, &event, r.secretKey); err != nil {
				log.Printf("error decrypting event: %v\n", err)
			}
			r.hub.broadcast <- &event
		case redis.Subscription:
			log.Printf("subscription %s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			return v
		}
	}
}

// Get value.
func (r *Redis) Get(ctx context.Context, k string) (string, error) {
	redisConn := r.redisPool.Get()
	defer redisConn.Close()
	s, err := redis.String(redisConn.Do("GET", k))
	if err == redis.ErrNil {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return s, nil
}

// Expire value.
func (r *Redis) Expire(ctx context.Context, k string, dt time.Duration) error {
	redisConn := r.redisPool.Get()
	defer redisConn.Close()
	seconds := int64(dt / time.Second)
	if _, err := redisConn.Do("EXPIRE", k, seconds); err != nil {
		return err
	}
	return nil
}

// Increment value.
func (r *Redis) Increment(ctx context.Context, k string) (int64, error) {
	redisConn := r.redisPool.Get()
	defer redisConn.Close()
	n, err := redis.Int64(redisConn.Do("INCR", k))
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Set value.
func (r *Redis) Set(ctx context.Context, k string, v string) error {
	redisConn := r.redisPool.Get()
	defer redisConn.Close()
	if _, err := redisConn.Do("SET", k, v); err != nil {
		return err
	}
	return nil
}

// Delete value.
func (r *Redis) Delete(ctx context.Context, k string) error {
	redisConn := r.redisPool.Get()
	defer redisConn.Close()
	if _, err := redisConn.Do("DEL", k); err != nil {
		return err
	}
	return nil
}
