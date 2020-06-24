package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/ds"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// VaultChange describes a vault item at a point in time.
type VaultChange struct {
	// Path for change /{collection}/{id}.
	Path string `msgpack:"p"`
	// Data ...
	Data []byte `msgpack:"dat"`
	// Nonce to prevent replay.
	Nonce string `msgpack:"n"`

	// Version is set by clients from remote change API.
	// This is untrusted.
	Version int64 `msgpack:"v,omitempty"`
	// Timestamp is set by clients from the remote change API.
	// This is untrusted.
	Timestamp time.Time `msgpack:"ts,omitempty"`
}

// Vault changes from the remote, decrypted with vault key.
type Vault struct {
	Changes []*VaultChange
	Version string
}

// VaultChanged saves vault changes to the remote.
// The changes are encrypted with the vault key before being sent to the remote.
func (c *Client) VaultChanged(ctx context.Context, key *keys.EdX25519Key, changes []*VaultChange) error {
	path := ds.Path("vault", key.ID())
	vals := url.Values{}

	out := []*api.VaultBox{}
	for _, change := range changes {
		if !change.Timestamp.IsZero() {
			return errors.Errorf("timestamp shouldn't be set for vault save")
		}
		if change.Nonce == "" {
			return errors.Errorf("nonce isn't set")
		}
		b, err := msgpack.Marshal(change)
		if err != nil {
			return err
		}
		out = append(out, &api.VaultBox{
			Data: vaultEncrypt(b, key),
		})
	}

	b, err := json.Marshal(out)
	if err != nil {
		return err
	}

	if _, err := c.putDocument(ctx, path, vals, key, bytes.NewReader(b)); err != nil {
		return err
	}
	return nil
}

// VaultOptions options for Vault.
type VaultOptions struct {
	// Version to list to/from
	Version string
	// Limit by
	Limit int
}

// VaultOption option.
type VaultOption func(o *VaultOptions)

// VaultVersion ...
func VaultVersion(version string) VaultOption {
	return func(o *VaultOptions) {
		o.Version = version
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

// Vault changes.
// Vault data is decrypted using the vault key before being returned.
// Clients should check for repeated nonces.
func (c *Client) Vault(ctx context.Context, key *keys.EdX25519Key, opt ...VaultOption) (*Vault, error) {
	opts := newVaultOptions(opt...)
	path := ds.Path("vault", key.ID())
	params := url.Values{}
	if opts.Version != "" {
		params.Add("version", opts.Version)
	}
	if opts.Limit != 0 {
		params.Add("limit", fmt.Sprintf("%d", opts.Limit))
	}

	// TODO: What if we hit limit, we won't have all the items

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
	out := make([]*VaultChange, 0, len(resp.Boxes))
	for _, box := range resp.Boxes {
		decrypted, err := vaultDecrypt(box.Data, key)
		if err != nil {
			return nil, err
		}
		var chg VaultChange
		if err := msgpack.Unmarshal(decrypted, &chg); err != nil {
			return nil, err
		}
		chg.Timestamp = box.Timestamp
		chg.Version = box.Version
		out = append(out, &chg)
	}
	return &Vault{Changes: out, Version: resp.Version}, nil
}

func vaultEncrypt(b []byte, key *keys.EdX25519Key) []byte {
	return keys.BoxSeal(b, key.X25519Key().PublicKey(), key.X25519Key())
}

func vaultDecrypt(b []byte, key *keys.EdX25519Key) ([]byte, error) {
	return keys.BoxOpen(b, key.X25519Key().PublicKey(), key.X25519Key())
}
