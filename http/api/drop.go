package api

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
)

// Drop ...
type Drop struct {
	Type   DropType `json:"type" msgpack:"type"`
	Sender keys.ID  `json:"sender" msgpack:"sender"`

	// ChannelDrop
	// Key
	Key *api.Key `json:"key,omitempty" msgpack:"key,omitempty"`

	// TokenDrop
	// Token
	Token string `json:"token,omitempty" msgpack:"token,omitempty"`
}

// DropType ...
type DropType string

// Drop types.
const (
	ChannelDrop DropType = "channel"
	TokenDrop   DropType = "token"
)

// NewChannelDrop creates a channel drop.
func NewChannelDrop(channel *keys.EdX25519Key, sender keys.ID) *Drop {
	return &Drop{
		Type:   ChannelDrop,
		Key:    api.NewKey(channel),
		Sender: sender,
	}
}

// NewTokenDrop creates a token drop.
func NewTokenDrop(token string, sender keys.ID) *Drop {
	return &Drop{
		Type:   TokenDrop,
		Token:  token,
		Sender: sender,
	}
}

// DecryptDrop decrypts a drop.
func DecryptDrop(b []byte, kr saltpack.Keyring) (*Drop, error) {
	var drop Drop
	pk, err := Decrypt(b, &drop, kr)
	if err != nil {
		return nil, err
	}
	if !keys.X25519Match(pk, drop.Sender) {
		return nil, errors.Errorf("drop sender mismatch")
	}
	return &drop, nil
}

// DropsResponse ...
type DropsResponse struct {
	Drops []*events.Event `json:"drops" msgpack:"drops"`
}
