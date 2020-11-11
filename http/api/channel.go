package api

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/saltpack"
)

// Channel ...
type Channel struct {
	ID keys.ID `json:"id" msgpack:"id"`
}

// ChannelInvite provides an encrypted key to a recipient.
type ChannelInvite struct {
	Channel      keys.ID `json:"channel" msgpack:"channel"`
	Recipient    keys.ID `json:"recipient" msgpack:"recipient"`
	Sender       keys.ID `json:"member" msgpack:"member"`
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
	sk, err := key.AsEdX25519()
	if err != nil {
		return nil, "", err
	}
	return sk, from, nil
}

// ChannelInfo is encrypted by clients.
type ChannelInfo struct {
	Channel keys.ID `json:"channel" msgpack:"channel"`
	Name    string  `json:"name,omitempty" msgpack:"name,omitempty"`

	Sender keys.ID `json:"-" msgpack:"-"` // Sender set by decryption
}

// ChannelMember ...
type ChannelMember struct {
	Member  keys.ID `json:"member" msgpack:"member"`
	Channel keys.ID `json:"channel" msgpack:"channel"`
	From    keys.ID `json:"from" msgpack:"from"`
}

// ChannelInvitesResponse ...
type ChannelInvitesResponse struct {
	Invites []*ChannelInvite `json:"invites" msgpack:"invites"`
}

// ChannelMembersResponse ..
type ChannelMembersResponse struct {
	Members []*ChannelMember `json:"members" msgpack:"members"`
}

// ChannelMembersAddRequest ...
// type ChannelMembersAddRequest struct {
// 	Members []*ChannelMember `json:"members" msgpack:"members"`
// }
