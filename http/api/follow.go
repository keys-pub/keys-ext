package api

import "github.com/keys-pub/keys"

// Follow user from kid.
type Follow struct {
	KID  keys.ID `json:"kid" msgpack:"kid"`
	User keys.ID `json:"user" msgpack:"user"`
}

// FollowsResponse ...
type FollowsResponse struct {
	Follows []*Follow `json:"follows" msgpack:"follows"`
}
