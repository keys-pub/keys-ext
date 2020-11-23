package server

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/keys-pub/keys/tsutil"
)

// Redis defines interface for a distributed key value store.
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
	// Publish data to key.
	Publish(ctx context.Context, k string, b []byte) error
}

type rdsTest struct {
	sync.Mutex
	kv    map[string]*rdsEntry
	clock tsutil.Clock
}

// NewRedisTest returns Redis for testing.
func NewRedisTest(clock tsutil.Clock) Redis {
	return newRedis(clock)
}

func newRedis(clock tsutil.Clock) *rdsTest {
	return &rdsTest{
		kv:    map[string]*rdsEntry{},
		clock: clock,
	}
}

type rdsEntry struct {
	Value  string
	Expire time.Time
}

func (m *rdsTest) Get(ctx context.Context, k string) (string, error) {
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

func (m *rdsTest) get(ctx context.Context, k string) (*rdsEntry, error) {
	e, ok := m.kv[k]
	if !ok {
		return nil, nil
	}
	if e.Expire.IsZero() {
		return e, nil
	}
	now := m.clock.Now()
	if e.Expire.Equal(now) || now.After(e.Expire) {
		return nil, nil
	}
	return e, nil
}

func (m *rdsTest) Expire(ctx context.Context, k string, dt time.Duration) error {
	m.Lock()
	defer m.Unlock()
	t := m.clock.Now()
	t = t.Add(dt)
	e, err := m.get(ctx, k)
	if err != nil {
		return err
	}
	e.Expire = t
	return m.set(ctx, k, e)
}

func (m *rdsTest) Delete(ctx context.Context, k string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.kv, k)
	return nil
}

func (m *rdsTest) Set(ctx context.Context, k string, v string) error {
	m.Lock()
	defer m.Unlock()
	return m.set(ctx, k, &rdsEntry{Value: v})
}

func (m *rdsTest) set(ctx context.Context, k string, e *rdsEntry) error {
	m.kv[k] = e
	return nil
}

func (m *rdsTest) Increment(ctx context.Context, k string) (int64, error) {
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

func (m *rdsTest) Publish(ctx context.Context, k string, b []byte) error {
	// TODO: Assert published events
	return nil
}
