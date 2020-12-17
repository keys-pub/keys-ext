package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// Vault events from the API, decrypted with vault API key.
type Vault struct {
	Events    []*VaultEvent
	Index     int64
	Truncated bool
}

// VaultEvent describes a vault event.
type VaultEvent struct {
	// Path for event.
	Path string `json:"path" msgpack:"p"`
	// Data ...
	Data []byte `json:"data" msgpack:"dat"`

	// RemoteIndex is set from the remote events API (untrusted).
	RemoteIndex int64 `json:"-" msgpack:"-"`
	// RemoteTimestamp is set from the remote events API (untrusted).
	RemoteTimestamp time.Time `json:"-" msgpack:"-"`

	// Deprecated fields, don't reuse these tag names.
	// Nonce to prevent replay.
	// Nonce []byte `msgpack:"n"`
	// Prev is a hash of the previous item.
	// Prev []byte `msgpack:"prv,omitempty"`
}

// VaultSend saves events to the vault API with a key.
// Events are encrypted with the key before saving.
func (c *Client) VaultSend(ctx context.Context, key *keys.EdX25519Key, events []*VaultEvent) error {
	path := dstore.Path("vault", key.ID())
	vals := url.Values{}

	out := []*api.Data{}
	for _, event := range events {
		if !event.RemoteTimestamp.IsZero() {
			return errors.Errorf("remote timestamp should be omitted on send")
		}
		if event.RemoteIndex != 0 {
			return errors.Errorf("remote index should be omitted on send")
		}
		b, err := msgpack.Marshal(event)
		if err != nil {
			return err
		}
		out = append(out, &api.Data{
			Data: vaultEncrypt(b, key),
		})
	}

	// TODO: Support msgpack
	b, err := json.Marshal(out)
	if err != nil {
		return err
	}

	if _, err := c.post(ctx, path, vals, bytes.NewReader(b), http.ContentHash(b), key); err != nil {
		return err
	}
	return nil
}

// VaultOptions options for Vault.
type VaultOptions struct {
	// Index to list to/from
	Index int64
	// Limit by
	Limit int
}

// VaultOption option.
type VaultOption func(o *VaultOptions)

// VaultIndex ...
func VaultIndex(index int64) VaultOption {
	return func(o *VaultOptions) {
		o.Index = index
	}
}

// VaultLimit ...
func VaultLimit(limit int) VaultOption {
	return func(o *VaultOptions) {
		o.Limit = limit
	}
}

func newVaultOptions(opts ...VaultOption) VaultOptions {
	var options VaultOptions
	for _, o := range opts {
		o(&options)
	}
	return options
}

// Vault events.
// Vault data is decrypted using the vault key before being returned.
// If truncated, there are more results if you call again with the new index.
func (c *Client) Vault(ctx context.Context, key *keys.EdX25519Key, opt ...VaultOption) (*Vault, error) {
	opts := newVaultOptions(opt...)
	path := dstore.Path("vault", key.ID())
	params := url.Values{}
	if opts.Index != 0 {
		params.Add("idx", strconv.FormatInt(opts.Index, 10))
	}
	if opts.Limit != 0 {
		// TODO: Support limit
		return nil, errors.Errorf("limit not currently supported")
	}

	resp, err := c.get(ctx, path, params, key)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	var out api.VaultResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}

	return vaultDecryptResponse(&out, key)
}

func vaultDecryptResponse(resp *api.VaultResponse, key *keys.EdX25519Key) (*Vault, error) {
	out := make([]*VaultEvent, 0, len(resp.Vault))
	for _, revent := range resp.Vault {
		decrypted, err := vaultDecrypt(revent.Data, key)
		if err != nil {
			return nil, err
		}
		var event VaultEvent
		if err := msgpack.Unmarshal(decrypted, &event); err != nil {
			return nil, err
		}
		event.RemoteTimestamp = tsutil.ConvertMillis(revent.Timestamp)
		event.RemoteIndex = revent.Index
		out = append(out, &event)
	}
	return &Vault{Events: out, Index: resp.Index, Truncated: resp.Truncated}, nil
}

func vaultEncrypt(b []byte, key *keys.EdX25519Key) []byte {
	return keys.BoxSeal(b, key.X25519Key().PublicKey(), key.X25519Key())
}

func vaultDecrypt(b []byte, key *keys.EdX25519Key) ([]byte, error) {
	return keys.BoxOpen(b, key.X25519Key().PublicKey(), key.X25519Key())
}

// VaultDelete removes a vault.
func (c *Client) VaultDelete(ctx context.Context, key *keys.EdX25519Key) error {
	path := dstore.Path("vault", key.ID())
	vals := url.Values{}

	if _, err := c.delete(ctx, path, vals, nil, "", key); err != nil {
		return err
	}
	return nil
}

// VaultExists checks if vault exists.
func (c *Client) VaultExists(ctx context.Context, key *keys.EdX25519Key) (bool, error) {
	path := dstore.Path("vault", key.ID())
	params := url.Values{}
	resp, err := c.head(ctx, path, params, key)
	if err != nil {
		return false, err
	}
	if resp == nil {
		return false, nil
	}

	return true, nil
}

// NewVaultEvent creates a new event.
func NewVaultEvent(path string, b []byte) *VaultEvent {
	return &VaultEvent{
		Path: path,
		Data: b,
	}
}
