package server

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// MemCache defines interface for memcache.
// Used to prevent nonce re-use for authenticated requests.
type MemCache interface {
	// Get returns value at key.
	Get(ctx context.Context, k string) (string, error)
	// Put puts a value at key.
	Set(ctx context.Context, k string, v string) error
	// Expire key.
	Expire(ctx context.Context, k string, dt time.Duration) error
	// Increment value at key.
	Increment(ctx context.Context, k string) (int64, error)
	// Publish value to key.
	Publish(ctx context.Context, k string, v string) error
	// Subscribe to key.
	Subscribe(ctx context.Context, k string, ch chan []byte) error
}

type memCache struct {
	sync.Mutex
	kv    map[string]*mcEntry
	nowFn func() time.Time

	pubsub map[string]chan []byte
}

// NewMemTestCache returns in memory MemCache (for testing).
func NewMemTestCache(nowFn func() time.Time) MemCache {
	return newMemTestCache(nowFn)
}

func newMemTestCache(nowFn func() time.Time) *memCache {
	kv := map[string]*mcEntry{}
	pubsub := map[string]chan []byte{}
	return &memCache{
		kv:     kv,
		nowFn:  nowFn,
		pubsub: pubsub,
	}
}

type mcEntry struct {
	Value  string
	Expire time.Time
}

func (m *memCache) Get(ctx context.Context, k string) (string, error) {
	m.Lock()
	defer m.Unlock()
	e, err := m.get(ctx, k)
	if err != nil {
		return "", err
	}
	if e == nil {
		return "", nil
	}
	return e.Value, nil
}

func (m *memCache) get(ctx context.Context, k string) (*mcEntry, error) {
	e, ok := m.kv[keys.Path("memcache", k)]
	if !ok {
		return nil, nil
	}
	if e.Expire.IsZero() {
		return e, nil
	}
	now := m.nowFn()
	if e.Expire.Equal(now) || now.After(e.Expire) {
		return nil, nil
	}
	return e, nil
}

func (m *memCache) Expire(ctx context.Context, k string, dt time.Duration) error {
	m.Lock()
	defer m.Unlock()
	t := m.nowFn()
	t = t.Add(dt)
	e, err := m.get(ctx, k)
	if err != nil {
		return err
	}
	e.Expire = t
	return m.set(ctx, k, e)
}

func (m *memCache) Set(ctx context.Context, k string, v string) error {
	m.Lock()
	defer m.Unlock()
	return m.set(ctx, k, &mcEntry{Value: v})
}

func (m *memCache) set(ctx context.Context, k string, e *mcEntry) error {
	m.kv[keys.Path("memcache", k)] = e
	return nil
}

func (m *memCache) Increment(ctx context.Context, k string) (int64, error) {
	m.Lock()
	defer m.Unlock()
	e, err := m.get(ctx, k)
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseInt(e.Value, 10, 64)
	if err != nil {
		return 0, err
	}
	n++
	inc := strconv.FormatInt(n, 10)
	e.Value = inc
	return n, m.set(ctx, k, e)
}

func (m *memCache) Publish(ctx context.Context, k string, v string) error {
	ch, ok := m.pubsub[k]
	if !ok {
		return errors.Errorf("no subscribe for %s", k)
	}
	logger.Debugf(ctx, "Publishing bytes to channel for %s", k)
	ch <- []byte(v)
	return nil
}

func (m *memCache) Subscribe(ctx context.Context, k string, ch chan []byte) error {
	m.pubsub[k] = ch
	return nil
}
