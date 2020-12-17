package api

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
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

// ChannelJoin is a user joined a channel (invite was accepted) (notification).
type ChannelJoin struct {
	User keys.ID `json:"user" msgpack:"user"`
}

// ChannelLeave if a user left the channel (notification).
type ChannelLeave struct {
	User keys.ID `json:"user" msgpack:"user"`
}

// ChannelInvite if invited to a channel.
type ChannelInvite struct {
	Channel   keys.ID      `json:"channelId" msgpack:"channelId"`
	Recipient keys.ID      `json:"recipient" msgpack:"recipient"`
	Key       *api.Key     `json:"key" msgpack:"key"`
	Info      *ChannelInfo `json:"info" msgpack:"info"`
}
