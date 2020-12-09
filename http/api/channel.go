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
	Channel   keys.ID `json:"channel" msgpack:"channel"`
	Recipient keys.ID `json:"recipient" msgpack:"recipient"`
	Key       []byte  `json:"k" msgpack:"k"`       // Encrypted api.Key to recipient from sender.
	Info      []byte  `json:"info" msgpack:"info"` // Encrypted api.ChannelInfo to recipient from sender.
}

// NewChannelInvite creates a channel invite.
func NewChannelInvite(channel *keys.EdX25519Key, info *ChannelInfo, sender *keys.EdX25519Key, recipient keys.ID) (*ChannelInvite, error) {
	ek, err := api.EncryptKey(api.NewKey(channel), sender, recipient, false)
	if err != nil {
		return nil, err
	}
	ei, err := Encrypt(info, sender, recipient)
	if err != nil {
		return nil, err
	}
	return &ChannelInvite{
		Channel:   channel.ID(),
		Recipient: recipient,
		Key:       ek,
		Info:      ei,
	}, nil
}

// DecryptKey for recipient keyring.
func (i *ChannelInvite) DecryptKey(kr saltpack.Keyring) (*keys.EdX25519Key, keys.ID, error) {
	key, sender, err := api.DecryptKey(i.Key, kr, false)
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

// DecryptInfo for recipient keyring.
func (i *ChannelInvite) DecryptInfo(kr saltpack.Keyring) (*ChannelInfo, keys.ID, error) {
	var info ChannelInfo
	pk, err := Decrypt(i.Info, &info, kr)
	if err != nil {
		return nil, "", err
	}
	return &info, pk, nil
}

// ChannelUser ...
type ChannelUser struct {
	Channel keys.ID `json:"channel" msgpack:"channel"`
	User    keys.ID `json:"user" msgpack:"user"`
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

// ChannelUninviteRequest ...
type ChannelUninviteRequest struct {
	// Message to post on uninvite.
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

// ChannelInvites if invites were sent (notification).
type ChannelInvites struct {
	Users []keys.ID `json:"users" msgpack:"users"`
}

// ChannelUninvites if invites were removed (notification).
type ChannelUninvites struct {
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

// UserChannelsResponse ...
type UserChannelsResponse struct {
	Channels []*Channel `json:"channels"`
}

// UserChannelInviteResponse ...
type UserChannelInviteResponse struct {
	Invite *ChannelInvite `json:"invite"`
}

// UserChannelInvitesResponse ...
type UserChannelInvitesResponse struct {
	Invites []*ChannelInvite `json:"invites"`
}
