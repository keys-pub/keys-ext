package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/ds"
)

// VaultSave saves to the vault.
func (c *Client) VaultSave(ctx context.Context, key *keys.EdX25519Key, b []byte) error {
	path := ds.Path("vault", key.ID())
	vals := url.Values{}
	_, err := c.postDocument(ctx, path, vals, key, bytes.NewReader(b))
	if err != nil {
		return err
	}

	return nil
}

// VaultOpts options for Vault.
type VaultOpts struct {
	// Version to list to/from
	Version string
	// Limit by
	Limit int
}

// Vault returns vault items.
func (c *Client) Vault(ctx context.Context, key *keys.EdX25519Key, opts *VaultOpts) (*api.VaultResponse, error) {
	if opts == nil {
		opts = &VaultOpts{}
	}
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

	return &resp, nil
}
