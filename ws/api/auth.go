package api

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
)

// GenerateAuth creates a new auth statement for key and url.
func GenerateAuth(key *keys.EdX25519Key, url string) []byte {
	st := &keys.Statement{
		KID:       key.ID(),
		Nonce:     keys.Rand16()[:],
		Timestamp: time.Now(),
		Type:      "auth@" + url,
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

// NonceCheck checks nonces.
type NonceCheck func(ctx context.Context, nonce string) error

// Authorize auth data from client.
// Returns verified key ID.
func Authorize(ctx context.Context, b []byte, now time.Time, url string, nonceCheck NonceCheck) (keys.ID, error) {
	var st keys.Statement
	if err := json.Unmarshal(b, &st); err != nil {
		return "", err
	}
	if err := AuthorizeStatement(ctx, &st, now, url, nonceCheck); err != nil {
		return "", err
	}
	return st.KID, nil
}

// AuthorizeStatement from client.
// Returns an error if it fails to verify.
func AuthorizeStatement(ctx context.Context, st *keys.Statement, now time.Time, url string, nonceCheck NonceCheck) error {
	if err := st.Verify(); err != nil {
		return err
	}

	// Check host
	expected := "auth@" + url
	if st.Type != expected {
		return errors.Errorf("auth invalid %q != %q", st.Type, expected)
	}

	// Check timestamp
	diff := now.Sub(st.Timestamp)
	if diff < 0 {
		diff = diff * -1
	}
	if diff > 30*time.Minute {
		return errors.Errorf("timestamp is invalid, diff %s", diff)
	}

	// Check nonce
	if nonceCheck == nil {
		return errors.Errorf("no nonce store set")
	}

	nonce := encoding.MustEncode(st.Nonce, encoding.Base62)

	if err := nonceCheck(ctx, nonce); err != nil {
		return err
	}

	return nil
}
