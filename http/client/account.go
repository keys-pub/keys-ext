package client

import (
	"context"
	"encoding/json"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/vault/auth"
)

func (c *Client) AccountCreate(ctx context.Context, account *keys.EdX25519Key, email string) error {
	path := dstore.Path("account", account.ID())
	create := &api.AccountCreateRequest{Email: email}
	body, err := json.Marshal(create)
	if err != nil {
		return err
	}
	if _, err := c.Request(ctx, &Request{Method: "PUT", Path: path, Body: body, Key: account}); err != nil {
		return err
	}
	return nil
}

func (c *Client) AccountAuthSave(ctx context.Context, account *keys.EdX25519Key, auth *auth.Auth) error {
	path := dstore.Path("account", account.ID(), "auths")

	encrypted, err := secretBoxMarshal(auth, account.Seed())
	if err != nil {
		return err
	}
	accountAuth := api.AccountAuth{
		ID:   auth.ID,
		Data: encrypted,
	}
	body, err := json.Marshal(accountAuth)
	if err != nil {
		return err
	}

	if _, err := c.Request(ctx, &Request{Method: "POST", Path: path, Body: body, Key: account}); err != nil {
		return err
	}
	return nil
}

func (c *Client) AccountAuths(ctx context.Context, account *keys.EdX25519Key) ([]*auth.Auth, error) {
	path := dstore.Path("account", account.ID(), "auths")
	resp, err := c.Request(ctx, &Request{Method: "GET", Path: path, Key: account})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	var out api.AccountAuthsResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}

	auths := []*auth.Auth{}
	for _, accountAuth := range out.Auths {
		var auth auth.Auth
		if err := secretBoxUnmarshal(accountAuth.Data, &auth, account.Seed()); err != nil {
			return nil, err
		}

		auths = append(auths, &auth)
	}

	return auths, nil
}
