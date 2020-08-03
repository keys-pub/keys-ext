package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/docs"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// Vault events from the API, decrypted with vault API key.
type Vault struct {
	Events []*Event
	Index  int64
}

// VaultSend saves events to the vault API with a key.
// Events are encrypted with the key before saving.
func (c *Client) VaultSend(ctx context.Context, key *keys.EdX25519Key, events []*Event) error {
	path := docs.Path("vault", key.ID())
	vals := url.Values{}

	out := []*api.Data{}
	for _, event := range events {
		if event.Timestamp != 0 {
			return errors.Errorf("timestamp shouldn't be set for vault send")
		}
		if len(event.Nonce) == 0 {
			return errors.Errorf("nonce isn't set")
		}
		b, err := msgpack.Marshal(event)
		if err != nil {
			return err
		}
		out = append(out, &api.Data{
			Data: vaultEncrypt(b, key),
		})
	}

	b, err := json.Marshal(out)
	if err != nil {
		return err
	}
	contentHash := api.ContentHash(b)

	if _, err := c.postDocument(ctx, path, vals, key, bytes.NewReader(b), contentHash); err != nil {
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
// Callers should check for repeated nonces and event chain ordering.
func (c *Client) Vault(ctx context.Context, key *keys.EdX25519Key, opt ...VaultOption) (*Vault, error) {
	opts := newVaultOptions(opt...)
	path := docs.Path("vault", key.ID())
	params := url.Values{}
	if opts.Index != 0 {
		params.Add("idx", strconv.FormatInt(opts.Index, 10))
	}
	if opts.Limit != 0 {
		// TODO: Support limit
		return nil, errors.Errorf("limit not currently supported")
	}

	doc, err := c.getDocument(ctx, path, params, key)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	var resp api.VaultResponse
	if err := json.Unmarshal(doc.Data, &resp); err != nil {
		return nil, err
	}

	return vaultDecryptResponse(&resp, key)
}

func vaultDecryptResponse(resp *api.VaultResponse, key *keys.EdX25519Key) (*Vault, error) {
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
		event.Timestamp = revent.Timestamp
		event.Index = revent.Index
		out = append(out, &event)
	}
	return &Vault{Events: out, Index: resp.Index}, nil
}

func vaultEncrypt(b []byte, key *keys.EdX25519Key) []byte {
	return keys.BoxSeal(b, key.X25519Key().PublicKey(), key.X25519Key())
}

func vaultDecrypt(b []byte, key *keys.EdX25519Key) ([]byte, error) {
	return keys.BoxOpen(b, key.X25519Key().PublicKey(), key.X25519Key())
}

// VaultDelete removes a vault.
func (c *Client) VaultDelete(ctx context.Context, key *keys.EdX25519Key) error {
	path := docs.Path("vault", key.ID())
	vals := url.Values{}

	if _, err := c.delete(ctx, path, vals, key); err != nil {
		return err
	}
	return nil
}

// VaultExists checks if vault exists.
func (c *Client) VaultExists(ctx context.Context, key *keys.EdX25519Key) (bool, error) {
	path := docs.Path("vault", key.ID())
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
