package api

import (
	"github.com/keys-pub/keys"
)

// Follow user.
type Follow struct {
	Sender    keys.ID `json:"sender" msgpack:"sender"`
	Recipient keys.ID `json:"recipient" msgpack:"recipient"`
}

// FollowResponse ...
type FollowResponse struct {
	Follow *Follow `json:"follow" msgpack:"follow"`
}

// FollowsResponse ...
type FollowsResponse struct {
	Follows []*Follow `json:"follows" msgpack:"follows"`
}
