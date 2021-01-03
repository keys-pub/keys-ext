package api

import "github.com/keys-pub/keys"

// DirectToken ...
type DirectToken struct {
	User  keys.ID `json:"user" msgpack:"user"`
	Token string  `json:"token" msgpack:"token"`
}
