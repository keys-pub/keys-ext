package server

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/keys-pub/keys/ds"
)

// Redis defines interface for memcache.
// Used to prevent nonce re-use for authenticated requests.
type Redis interface {
	// Get returns value at key.
	Get(ctx context.Context, k string) (string, error)
	// Put puts a value at key.
	Set(ctx context.Context, k string, v string) error
	// Delete key.
	Delete(ctx context.Context, k string) error
	// Expire key.
	Expire(ctx context.Context, k string, dt time.Duration) error
	// Increment value at key.
	Increment(ctx context.Context, k string) (int64, error)
}

type rds struct {
	sync.Mutex
	kv    map[string]*mcEntry
	nowFn func() time.Time
}

// NewRedisTest returns Redis for testing.
func NewRedisTest(nowFn func() time.Time) Redis {
	return newRedis(nowFn)
}

func newRedis(nowFn func() time.Time) *rds {
	kv := map[string]*mcEntry{}
	mc := &rds{
		kv:    kv,
		nowFn: nowFn,
	}
	return mc
}

type mcEntry struct {
	Value  string
	Expire time.Time
}

func (m *rds) Get(ctx context.Context, k string) (string, error) {
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

func (m *rds) get(ctx context.Context, k string) (*mcEntry, error) {
	e, ok := m.kv[ds.Path("memcache", k)]
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

func (m *rds) Expire(ctx context.Context, k string, dt time.Duration) error {
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

func (m *rds) Delete(ctx context.Context, k string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.kv, ds.Path("memcache", k))
	return nil
}

func (m *rds) Set(ctx context.Context, k string, v string) error {
	m.Lock()
	defer m.Unlock()
	return m.set(ctx, k, &mcEntry{Value: v})
}

func (m *rds) set(ctx context.Context, k string, e *mcEntry) error {
	m.kv[ds.Path("memcache", k)] = e
	return nil
}

func (m *rds) Increment(ctx context.Context, k string) (int64, error) {
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
