package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/docs"
	"github.com/pkg/errors"
)

// SigchainSave ...
func (c *Client) SigchainSave(ctx context.Context, st *keys.Statement) error {
	path := docs.Path("sigchain", st.URL())
	b, err := st.Bytes()
	if err != nil {
		return err
	}
	if _, err = c.put(ctx, path, url.Values{}, nil, bytes.NewReader(b)); err != nil {
		return err
	}
	return nil
}

// Sigchain for KID. If sigchain not found, a nil response is returned.
func (c *Client) Sigchain(ctx context.Context, kid keys.ID) (*api.SigchainResponse, error) {
	path := docs.Path("sigchain", kid)

	params := url.Values{}
	params.Add("include", "md")
	doc, err := c.getDocument(ctx, path, params, nil)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	var resp api.SigchainResponse
	if err := json.Unmarshal(doc.Data, &resp); err != nil {
		return nil, err
	}

	if resp.KID != kid {
		return nil, errors.Errorf("mismatched id in response %q != %q", resp.KID, kid)
	}

	return &resp, nil
}
