package client

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/keys-pub/keys"
)

// CreateUser creates a user.
func (c *Client) CreateUser(ctx context.Context, user *User) (*User, error) {
	bin, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	resp, err := c.Post(ctx, "/api/v4/users", nil, bytes.NewReader(bin))
	if err != nil {
		return nil, err
	}

	var out User
	if err := unmarshal(resp, &out); err != nil {
		return nil, err
	}
	return &out, err
}

// CreateUserWithKey creates a user with a key.
func (c *Client) CreateUserWithKey(ctx context.Context, key *keys.EdX25519Key) (*User, error) {
	user := &User{
		ID:       key.ID().String(),
		Username: key.ID().String(),
		Password: keys.RandPassword(16),
		Email:    key.ID().String() + "@keys.pub",
	}
	return c.CreateUser(ctx, user)
}
