package client

import (
	"crypto/sha256"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/encoding"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// Event describes a client event.
type Event struct {
	// Path for event /{collection}/{id}.
	Path string `msgpack:"p"`
	// Data ...
	Data []byte `msgpack:"dat"`

	// Nonce to prevent replay.
	Nonce []byte `msgpack:"n"`
	// Prev is a hash of the previous item (optional if root).
	Prev []byte `msgpack:"prv,omitempty"`

	// Index is set by clients from remote events API (untrusted).
	Index int64 `msgpack:"idx,omitempty"`
	// Timestamp is set by clients from the remote events API (untrusted).
	Timestamp time.Time `msgpack:"ts,omitempty"`
}

// NewEvent creates a new event.
func NewEvent(path string, b []byte, prev *Event) *Event {
	var phash []byte
	if prev != nil {
		phash = EventHash(prev)[:]
	}
	return &Event{
		Path:  path,
		Data:  b,
		Nonce: keys.RandBytes(24),
		Prev:  phash,
	}
}

// EventHash returns hash for Event.
func EventHash(event *Event) *[32]byte {
	b, err := msgpack.Marshal(event)
	if err != nil {
		panic(err)
	}
	h := sha256.Sum256(b)
	return &h
}

// CheckEventchain checks event previous hashes.
func CheckEventchain(events []*Event) error {
	set := docs.NewStringSet()
	for i, event := range events {
		if event.Prev == nil {
			if i != 0 {
				return errors.Errorf("previous event hash is nil")
			}
			set.Add(encoding.EncodeBase64(EventHash(event)[:]))
			continue
		}
		if !set.Contains(encoding.EncodeBase64(event.Prev)) {
			return errors.Errorf("previous event hash not found")
		}
		set.Add(encoding.EncodeBase64(EventHash(event)[:]))
	}
	return nil
}
