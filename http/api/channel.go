package api

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
)

// Channel ...
type Channel struct {
	ID keys.ID `json:"id" msgpack:"id"`

	Index     int64 `json:"idx,omitempty" msgpack:"idx,omitempty"`
	Timestamp int64 `json:"ts,omitempty" msgpack:"ts,omitempty"`
}

// ChannelInvite provides an encrypted key to a recipient.
type ChannelInvite struct {
	Channel      keys.ID `json:"channel" msgpack:"channel"`
	Recipient    keys.ID `json:"recipient" msgpack:"recipient"`
	Sender       keys.ID `json:"sender" msgpack:"sender"`
	EncryptedKey []byte  `json:"k" msgpack:"k"` // Encrypted api.Key to recipient
}

// Key decrypted by recipient.
func (i *ChannelInvite) Key(recipient *keys.EdX25519Key) (*keys.EdX25519Key, keys.ID, error) {
	key, sender, err := api.DecryptKey(i.EncryptedKey, saltpack.NewKeyring(recipient))
	if err != nil {
		return nil, "", err
	}
	var from keys.ID
	if sender != nil {
		from = sender.ID()
	}
	sk := key.AsEdX25519()
	if sk == nil {
		return nil, "", errors.Errorf("invalid key")
	}
	return sk, from, nil
}

// ChannelUser ...
type ChannelUser struct {
	Channel keys.ID `json:"channel" msgpack:"channel"`
	User    keys.ID `json:"user" msgpack:"user"`
	From    keys.ID `json:"from" msgpack:"from"`
}

// ChannelCreateRequest ...
type ChannelCreateRequest struct {
	// Message to post on create.
	Message []byte `json:"msg,omitempty" msgpack:"msg,omitempty"`
}

// ChannelInvitesRequest ...
type ChannelInvitesRequest struct {
	Invites []*ChannelInvite `json:"invites" msgpack:"invites"`
	// Message to post on invite.
	Message []byte `json:"msg,omitempty" msgpack:"msg,omitempty"`
}

// ChannelInvitesResponse ...
type ChannelInvitesResponse struct {
	Invites []*ChannelInvite `json:"invites" msgpack:"invites"`
}

// ChannelUsersResponse ..
type ChannelUsersResponse struct {
	Users []*ChannelUser `json:"users" msgpack:"users"`
}

// ChannelJoinRequest ...
type ChannelJoinRequest struct {
	// Message to post on join.
	Message []byte `json:"msg,omitempty" msgpack:"msg,omitempty"`
}

// ChannelLeaveRequest ...
type ChannelLeaveRequest struct {
	// Message to post on leave.
	Message []byte `json:"msg,omitempty" msgpack:"msg,omitempty"`
}

// ChannelUsersAddRequest ...
// type ChannelUsersAddRequest struct {
// 	Users []*ChannelUser `json:"users" msgpack:"users"`
// }

// ChannelInfo for setting channel name or description.
type ChannelInfo struct {
	Name        string `json:"name,omitempty" msgpack:"name,omitempty"`
	Description string `json:"desc,omitempty" msgpack:"desc,omitempty"`
}

// ChannelInvitesNn if invites were sent (notification).
type ChannelInvitesNn struct {
	Users []keys.ID `json:"users" msgpack:"users"`
}

// ChannelJoinNn is a user joined a channel (invite was accepted) (notification).
type ChannelJoinNn struct {
	User keys.ID `json:"user" msgpack:"user"`
}

// ChannelLeaveNn if a user left the channel (notification).
type ChannelLeaveNn struct {
	User keys.ID `json:"user" msgpack:"user"`
}
