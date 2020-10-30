package api

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
)

// GenerateAuth creates a new auth statement.
func GenerateAuth(key *keys.EdX25519Key, host string) []byte {
	st := &keys.Statement{
		KID:       key.ID(),
		Nonce:     keys.Rand16()[:],
		Timestamp: time.Now(),
		Data:      []byte(host),
		Type:      "auth",
	}
	if err := st.Sign(key); err != nil {
		panic(err)
	}
	b, err := json.Marshal(st)
	if err != nil {
		panic(err)
	}
	return b
}

// Nonces defines interface for a nonce store.
// Used to prevent nonce re-use for authenticated requests.
type Nonces interface {
	// Get returns value at key.
	Get(ctx context.Context, k string) (string, error)
	// Put puts a value at key.
	Set(ctx context.Context, k string, v string) error
	// Delete key.
	Delete(ctx context.Context, k string) error
	// Expire key.
	Expire(ctx context.Context, k string, dt time.Duration) error
}

// CheckAuth ...
func CheckAuth(ctx context.Context, b []byte, now time.Time, host string, nonces Nonces) (keys.ID, error) {
	var st keys.Statement
	if err := json.Unmarshal(b, &st); err != nil {
		return "", err
	}
	if err := st.Verify(); err != nil {
		return "", err
	}

	// Check host
	if string(st.Data) != host {
		return "", errors.Errorf("auth host invalid %s != %s", string(st.Data), host)
	}

	// Check timestamp
	diff := now.Sub(st.Timestamp)
	if diff < 0 {
		diff = diff * -1
	}
	if diff > 30*time.Minute {
		return "", errors.Errorf("timestamp is invalid, diff %s", diff)
	}

	// Check nonce
	if nonces == nil {
		return "", errors.Errorf("no nonce store set")
	}
	if len(st.Nonce) == 0 {
		return "", errors.Errorf("missing nonce")
	}
	if len(st.Nonce) < 16 {
		return "", errors.Errorf("invalid nonce")
	}
	nonce := encoding.MustEncode(st.Nonce, encoding.Base62)

	val, err := nonces.Get(ctx, nonce)
	if err != nil {
		return "", err
	}
	if val != "" {
		return "", errors.Errorf("nonce collision")
	}
	if err := nonces.Set(ctx, nonce, "1"); err != nil {
		return "", err
	}
	if err := nonces.Expire(ctx, nonce, time.Hour); err != nil {
		return "", err
	}

	return st.KID, nil
}
