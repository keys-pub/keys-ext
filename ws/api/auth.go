package api

import (
	"encoding/json"
	"time"

	"github.com/keys-pub/keys"
)

// GenerateAuth creates a new auth statement.
func GenerateAuth(key *keys.EdX25519Key, urs string) []byte {
	st := &keys.Statement{
		KID:       key.ID(),
		Nonce:     keys.Rand24()[:],
		Timestamp: time.Now(),
		Data:      []byte(urs),
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

// CheckAuth ...
func CheckAuth(b []byte) (keys.ID, error) {
	var st keys.Statement
	if err := json.Unmarshal(b, &st); err != nil {
		return "", err
	}
	// TODO: Check nonce and timestamp and host (data)
	if err := st.Verify(); err != nil {
		return "", err
	}
	return st.KID, nil
}
