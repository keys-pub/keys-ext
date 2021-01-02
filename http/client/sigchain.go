package client

import (
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/pkg/errors"
)

// SigchainSave ...
func (c *Client) SigchainSave(ctx context.Context, st *keys.Statement) error {
	path := dstore.Path("sigchain", st.URL())
	b, err := st.Bytes()
	if err != nil {
		return err
	}
	_, err = c.retryOnConflict(ctx, request{Method: "PUT", Path: path, Body: b}, 1, 3, 2*time.Second)
	if err != nil {
		return err
	}

	return nil
}

// Sigchain for KID. If sigchain not found, a nil response is returned.
func (c *Client) Sigchain(ctx context.Context, kid keys.ID) (*api.SigchainResponse, error) {
	path := dstore.Path("sigchain", kid)

	params := url.Values{}
	params.Add("include", "md")
	resp, err := c.req(ctx, request{Method: "GET", Path: path, Params: params})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	var out api.SigchainResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}

	if out.KID != kid {
		return nil, errors.Errorf("mismatched id in response %q != %q", out.KID, kid)
	}

	return &out, nil
}
