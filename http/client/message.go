package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// MessageSend posts an encrypted message.
// TODO: expire time.Duration
func (c *Client) MessageSend(ctx context.Context, message *api.Message, sender *keys.EdX25519Key, channel *keys.EdX25519Key) error {
	// if expire == time.Duration(0) {
	// 	return errors.Errorf("no expire specified")
	// }
	if message.RemoteTimestamp != 0 {
		return errors.Errorf("remote timestamp should be omitted on send")
	}
	if message.RemoteIndex != 0 {
		return errors.Errorf("remote index should be omitted on send")
	}
	if message.Timestamp == 0 {
		return errors.Errorf("message timestamp is not set")
	}
	if message.Sender != "" && message.Sender != sender.ID() {
		return errors.Errorf("message sender mismatch")
	}

	encrypted, err := EncryptMessage(message, sender, channel.ID())
	if err != nil {
		return err
	}

	path := dstore.Path("channel", channel.ID(), "msgs")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", sender),
		http.NewAuthKey("Authorization-Channel", channel),
	)

	vals := url.Values{}
	// vals.Set("expire", expire.String())
	if _, err := c.post(ctx, path, vals, bytes.NewReader(encrypted), http.ContentHash(encrypted), auth); err != nil {
		return err
	}
	return nil
}

// MessagesOpts options for Messages.
type MessagesOpts struct {
	// Index to list to/from
	Index int64
	// Direction ascending or descending
	Direction events.Direction
	// Limit by
	Limit int
}

// Messages response.
type Messages struct {
	Messages  []*events.Event
	Index     int64
	Truncated bool
}

// Messages returns encrypted messages (as event.Event) and current index from a
// previous index.
// If truncated, there are more results if you call again with the new index.
// To decrypt to api.Message, use DecryptMessage.
func (c *Client) Messages(ctx context.Context, channel *keys.EdX25519Key, sender *keys.EdX25519Key, opts *MessagesOpts) (*Messages, error) {
	if opts == nil {
		opts = &MessagesOpts{}
	}

	path := dstore.Path("channel", channel.ID(), "msgs")
	params := url.Values{}
	if opts.Index != 0 {
		params.Add("idx", strconv.FormatInt(opts.Index, 10))
	}
	if opts.Direction != "" {
		params.Add("dir", string(opts.Direction))
	}
	if opts.Limit != 0 {
		params.Add("limit", fmt.Sprintf("%d", opts.Limit))
	}

	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", sender),
		http.NewAuthKey("Authorization-Channel", channel),
	)

	resp, err := c.get(ctx, path, params, auth)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	var out api.MessagesResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}

	return &Messages{
		Messages:  out.Messages,
		Index:     out.Index,
		Truncated: out.Truncated,
	}, nil
}

// EncryptMessage encrypts a message.
func EncryptMessage(message *api.Message, sender *keys.EdX25519Key, channel keys.ID) ([]byte, error) {
	b, err := msgpack.Marshal(message)
	if err != nil {
		return nil, err
	}
	encrypted, err := saltpack.Signcrypt(b, false, sender, channel.ID())
	if err != nil {
		return nil, err
	}
	return encrypted, nil
}

// DecryptMessage decrypts a remote Event from Messages.
func DecryptMessage(event *events.Event, kr saltpack.Keyring) (*api.Message, error) {
	decrypted, pk, err := saltpack.SigncryptOpen(event.Data, false, kr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decrypt message")
	}
	var message api.Message
	if err := msgpack.Unmarshal(decrypted, &message); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal message")
	}
	message.Sender = pk.ID()
	message.RemoteIndex = event.Index
	message.RemoteTimestamp = event.Timestamp
	return &message, nil
}
