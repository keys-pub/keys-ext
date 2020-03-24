package server

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/keys-pub/keys"
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
	Subscribe(ctx context.Context, k string) (<-chan []byte, error)
	// Unsubscribe to key.
	Unsubscribe(ctx context.Context, k string) error
}

type memCache struct {
	sync.Mutex
	kv    map[string]*mcEntry
	nowFn func() time.Time

	msgs   map[string][]string
	pubsub map[string]chan []byte
}

// NewMemTestCache returns in memory MemCache (for testing).
func NewMemTestCache(nowFn func() time.Time) MemCache {
	return newMemTestCache(nowFn)
}

func newMemTestCache(nowFn func() time.Time) *memCache {
	kv := map[string]*mcEntry{}
	pubsub := map[string]chan []byte{}
	msgs := map[string][]string{}
	mc := &memCache{
		kv:     kv,
		nowFn:  nowFn,
		pubsub: pubsub,
		msgs:   msgs,
	}
	return mc
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
	m.Lock()
	defer m.Unlock()
	vals, ok := m.msgs[k]
	if !ok {
		m.msgs[k] = []string{v}
	} else {
		vals = append(vals, v)
		m.msgs[k] = vals
	}

	m.pub(k)

	return nil
}

func (m *memCache) Subscribe(ctx context.Context, k string) (<-chan []byte, error) {
	m.Lock()
	defer m.Unlock()

	ch := make(chan []byte)

	m.pubsub[k] = ch

	m.pub(k)

	return ch, nil
}

func (m *memCache) Unsubscribe(ctx context.Context, k string) error {
	m.Lock()
	defer m.Unlock()
	ch, ok := m.pubsub[k]
	if !ok {
		return nil
	}
	close(ch)
	delete(m.msgs, k)
	delete(m.pubsub, k)
	return nil
}

func (m *memCache) pub(k string) {
	ch, ok := m.pubsub[k]
	if ok {
		vals, ok := m.msgs[k]
		if ok {
			delete(m.msgs, k)
			go func() {
				for _, v := range vals {
					ch <- []byte(v)
				}
			}()
		}
	}
}
