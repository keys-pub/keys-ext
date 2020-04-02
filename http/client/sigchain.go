package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/pkg/errors"
)

// PutSigchainStatement ...
func (c *Client) PutSigchainStatement(ctx context.Context, st *keys.Statement) error {
	path := keys.Path(st.URL())
	_, err := c.put(ctx, path, url.Values{}, nil, bytes.NewReader(st.Bytes()))
	return err
}

// Sigchain for KID. If sigchain not found, a nil response is returned.
func (c *Client) Sigchain(ctx context.Context, kid keys.ID) (*api.SigchainResponse, error) {
	path := keys.Path("sigchain", kid)

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
