package api

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
)

// UserToken ...
type UserToken struct {
	User  keys.ID `json:"user" msgpack:"user"`
	Token string  `json:"token" msgpack:"token"`
}

// UserTokenResponse ...
type UserTokenResponse struct {
	Token string `json:"token" msgpack:"token"`
}

// GenerateToken creates a token.
func GenerateToken() string {
	return encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
}
