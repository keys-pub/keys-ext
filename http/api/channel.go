package api

import (
	"github.com/keys-pub/keys"
)

// Channel ...
type Channel struct {
	ID keys.ID `json:"id" msgpack:"id"`

	Index     int64 `json:"idx,omitempty" msgpack:"idx,omitempty"`
	Timestamp int64 `json:"ts,omitempty" msgpack:"ts,omitempty"`
}

// ChannelCreateRequest ...
type ChannelCreateRequest struct {
	// Message to post on create.
	Message []byte `json:"msg,omitempty" msgpack:"msg,omitempty"`
}

// ChannelInfo for setting channel name or description.
type ChannelInfo struct {
	Name        string `json:"name,omitempty" msgpack:"name,omitempty"`
	Description string `json:"desc,omitempty" msgpack:"desc,omitempty"`
}

// ChannelInvites if invites were sent (notification).
type ChannelInvites struct {
	Users []keys.ID `json:"users" msgpack:"users"`
}

// ChannelJoin is a user joined a channel (invite was accepted) (notification).
type ChannelJoin struct {
	User keys.ID `json:"user" msgpack:"user"`
}

// ChannelLeave if a user left the channel (notification).
type ChannelLeave struct {
	User keys.ID `json:"user" msgpack:"user"`
}
