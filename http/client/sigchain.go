package client

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/pkg/errors"
)

// PutSigchainStatement ...
func (c *Client) PutSigchainStatement(st *keys.Statement) error {
	path := keys.Path(st.URL())
	_, err := c.put(path, url.Values{}, nil, bytes.NewReader(st.Bytes()))
	return err
}

// Sigchain for KID. If sigchain not found, a nil response is returned.
func (c *Client) Sigchain(kid keys.ID) (*api.SigchainResponse, error) {
	path := keys.Path("sigchain", kid)

	params := url.Values{}
	params.Add("include", "md")
	e, err := c.get(path, params, nil)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, nil
	}

	var resp api.SigchainResponse
	if err := json.Unmarshal(e.Data, &resp); err != nil {
		return nil, err
	}

	if resp.KID != kid {
		return nil, errors.Errorf("mismatched id in response %q != %q", resp.KID, kid)
	}

	return &resp, nil
}

// Sigchains ...
func (c *Client) Sigchains(version string) (*api.SigchainsResponse, error) {
	path := keys.Path("sigchains")

	params := url.Values{}
	params.Add("include", "md")
	params.Add("version", version)

	e, err := c.get(path, params, nil)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, errors.Errorf("sigchains response not found")
	}

	var val api.SigchainsResponse
	if err := json.Unmarshal(e.Data, &val); err != nil {
		return nil, err
	}
	return &val, nil
}
