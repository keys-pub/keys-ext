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
)

// VaultItem is data with timestamp from remote.
type VaultItem struct {
	// Data to be encrypted with vault key.
	Data []byte
	// Timestamp is set by remote.
	Timestamp time.Time
}

// Vault items from remote, decrypted.
type Vault struct {
	Items   []*VaultItem
	Version string
}

// VaultSave saves data into remote vault.
// The VaultItem data is encrypted with the vault key before being sent to the remote.
func (c *Client) VaultSave(ctx context.Context, key *keys.EdX25519Key, items []*VaultItem) error {
	path := ds.Path("vault", key.ID())
	vals := url.Values{}

	out := []*api.VaultItem{}
	for _, item := range items {
		if !item.Timestamp.IsZero() {
			return errors.Errorf("timestamp shouldn't be set for vault save")
		}
		out = append(out, &api.VaultItem{
			Data: vaultEncrypt(item.Data, key),
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
	out := make([]*VaultItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		b, err := vaultDecrypt(item.Data, key)
		if err != nil {
			return nil, err
		}
		out = append(out, &VaultItem{
			Data:      b,
			Timestamp: item.Timestamp,
		})
	}
	return &Vault{Items: out, Version: resp.Version}, nil
}

func vaultEncrypt(b []byte, key *keys.EdX25519Key) []byte {
	return keys.BoxSeal(b, key.X25519Key().PublicKey(), key.X25519Key())
}

func vaultDecrypt(b []byte, key *keys.EdX25519Key) ([]byte, error) {
	return keys.BoxOpen(b, key.X25519Key().PublicKey(), key.X25519Key())
}
