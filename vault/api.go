package vault

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// Client for API.
type Client struct {
	*client.Client
}

// NewClient creates a client.
func NewClient(client *client.Client) *Client {
	return &Client{client}
}

// Events events from the API, decrypted with vault API key.
type Events struct {
	Events    []*Event
	Index     int64
	Truncated bool
}

// Event describes a vault event.
type Event struct {
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
func (c *Client) VaultSend(ctx context.Context, key *keys.EdX25519Key, events []*Event) error {
	path := dstore.Path("vault", key.ID())

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

	if _, err := c.Request(ctx, &client.Request{Method: "POST", Path: path, Body: b, Key: key}); err != nil {
		return err
	}
	return nil
}

// Vault events.
// Vault data is decrypted using the vault key before being returned.
// If truncated, there are more results if you call again with the new index.
func (c *Client) Vault(ctx context.Context, key *keys.EdX25519Key, index int64) (*Events, error) {
	path := dstore.Path("vault", key.ID())
	params := url.Values{}
	if index != 0 {
		params.Add("idx", strconv.FormatInt(index, 10))
	}

	resp, err := c.Request(ctx, &client.Request{Method: "GET", Path: path, Params: params, Key: key})
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

func vaultDecryptResponse(resp *api.VaultResponse, key *keys.EdX25519Key) (*Events, error) {
	out := make([]*Event, 0, len(resp.Vault))
	for _, revent := range resp.Vault {
		decrypted, err := vaultDecrypt(revent.Data, key)
		if err != nil {
			return nil, err
		}
		var event Event
		if err := msgpack.Unmarshal(decrypted, &event); err != nil {
			return nil, err
		}
		event.RemoteTimestamp = tsutil.ParseMillis(revent.Timestamp)
		event.RemoteIndex = revent.Index
		out = append(out, &event)
	}
	return &Events{Events: out, Index: resp.Index, Truncated: resp.Truncated}, nil
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
	if _, err := c.Request(ctx, &client.Request{Method: "DELETE", Path: path, Key: key}); err != nil {
		return err
	}
	return nil
}

// VaultExists checks if vault exists.
func (c *Client) VaultExists(ctx context.Context, key *keys.EdX25519Key) (bool, error) {
	path := dstore.Path("vault", key.ID())
	params := url.Values{}
	resp, err := c.Request(ctx, &client.Request{Method: "HEAD", Path: path, Params: params, Key: key})
	if err != nil {
		return false, err
	}
	if resp == nil {
		return false, nil
	}

	return true, nil
}

// NewEvent creates a new event.
func NewEvent(path string, b []byte) *Event {
	return &Event{
		Path: path,
		Data: b,
	}
}
