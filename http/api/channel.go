package api

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
)

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
	Channel   keys.ID      `json:"channel" msgpack:"channel"`
	Recipient keys.ID      `json:"recipient" msgpack:"recipient"`
	Sender    keys.ID      `json:"sender" msgpack:"sender"`
	Key       *api.Key     `json:"key" msgpack:"key"`
	Token     string       `json:"token" msgpack:"token"`
	Info      *ChannelInfo `json:"info,omitempty" msgpack:"info,omitempty"`
}
