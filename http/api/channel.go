package api

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
)

// Channel ...
type Channel struct {
	ID keys.ID `json:"id" msgpack:"id"`

	Index     int64  `json:"idx,omitempty" msgpack:"idx,omitempty"`
	Timestamp int64  `json:"ts,omitempty" msgpack:"ts,omitempty"`
	Token     string `json:"token,omitempty" msgpack:"token,omitempty"`
}

// ChannelCreateRequest ...
type ChannelCreateRequest struct {
	// Message to post on create.
	Message []byte `json:"msg,omitempty" msgpack:"msg,omitempty"`
}

// ChannelCreateResponse ...
type ChannelCreateResponse struct {
	Channel *Channel `json:"channel,omitempty" msgpack:"channel,omitempty"`
}

// ChannelStatus ...
type ChannelStatus struct {
	ID        keys.ID `json:"id" msgpack:"id"`
	Index     int64   `json:"idx" msgpack:"idx"`
	Timestamp int64   `json:"ts" msgpack:"ts"`
}

// ChannelToken ...
type ChannelToken struct {
	ID    keys.ID
	Token string
}

// ChannelsStatusRequest ...
type ChannelsStatusRequest struct {
	Channels map[keys.ID]string `json:"channels,omitempty" msgpack:"channels,omitempty"`
}

// ChannelsStatusResponse ...
type ChannelsStatusResponse struct {
	Channels []*ChannelStatus `json:"channels,omitempty" msgpack:"channels,omitempty"`
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
	Channel   keys.ID      `json:"channel" msgpack:"channel"`
	Recipient keys.ID      `json:"recipient" msgpack:"recipient"`
	Sender    keys.ID      `json:"sender" msgpack:"sender"`
	Key       *api.Key     `json:"key" msgpack:"key"`
	Token     string       `json:"token" msgpack:"token"`
	Info      *ChannelInfo `json:"info,omitempty" msgpack:"info,omitempty"`
}
