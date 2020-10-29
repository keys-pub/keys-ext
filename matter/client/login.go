package client

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/http"
)

// LoginWithPassword logs a user in with a password.
func (c *Client) LoginWithPassword(ctx context.Context, username string, password string) (*User, error) {
	m := make(map[string]string)
	m["login_id"] = username
	m["password"] = password
	body, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	resp, err := c.Post(ctx, "/api/v4/users/login", nil, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.AuthToken = resp.Header.Get("token")
	c.AuthType = "BEARER"

	var user User
	if err := unmarshal(resp, &user); err != nil {
		return nil, err
	}
	return &user, err
}

// LoginWithKey logs a user in using an EdX25519 key.
func (c *Client) LoginWithKey(ctx context.Context, key *keys.EdX25519Key) (*User, error) {
	m := make(map[string]string)
	m["login_id"] = key.ID().String()
	body, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	contentHash := http.ContentHash([]byte(body))

	resp, err := c.PostWithKey(ctx, "/api/v4/users/login", nil, bytes.NewReader(body), key, contentHash)
	if err != nil {
		return nil, err
	}

	c.AuthToken = resp.Header.Get("token")
	c.AuthType = "BEARER"

	var user User
	if err := unmarshal(resp, &user); err != nil {
		return nil, err
	}
	return &user, err
}

// Logout clear auth token.
func (c *Client) Logout() {
	c.AuthType = ""
	c.AuthToken = ""
}
