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
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// MessageSend posts an encrypted message.
// TODO: expire time.Duration
func (c *Client) MessageSend(ctx context.Context, sender *keys.EdX25519Key, channel *keys.EdX25519Key, message *api.Message) error {
	// if expire == time.Duration(0) {
	// 	return errors.Errorf("no expire specified")
	// }
	if !message.RemoteTimestamp.IsZero() {
		return errors.Errorf("remote timestamp should be omitted on send")
	}
	if message.RemoteIndex != 0 {
		return errors.Errorf("remote index should be omitted on send")
	}
	if message.CreatedAt.IsZero() {
		return errors.Errorf("message.createdAt is not set")
	}
	if message.Sender != "" && message.Sender != sender.ID() {
		return errors.Errorf("message sender mismatch")
	}
	b, err := msgpack.Marshal(message)
	if err != nil {
		return err
	}
	encrypted, err := saltpack.Signcrypt(b, false, sender, channel.ID())
	if err != nil {
		return err
	}
	contentHash := http.ContentHash(encrypted)

	path := dstore.Path("channel", channel.ID(), "msgs")
	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", sender),
		http.NewAuthKey("Authorization-Channel", channel),
	)

	vals := url.Values{}
	// vals.Set("expire", expire.String())
	if _, err := c.post(ctx, path, vals, bytes.NewReader(encrypted), contentHash, auth); err != nil {
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

// Messages returns encrypted messages.
// To decrypt a message, use Client#MessageDecrypt.
func (c *Client) Messages(ctx context.Context, channel *keys.EdX25519Key, sender *keys.EdX25519Key, opts *MessagesOpts) ([]*events.Event, int64, error) {
	path := dstore.Path("channel", channel.ID(), "msgs")
	if opts == nil {
		opts = &MessagesOpts{}
	}

	params := url.Values{}
	params.Add("include", "md")
	if opts.Index != 0 {
		params.Add("idx", strconv.FormatInt(opts.Index, 10))
	}
	if opts.Direction != "" {
		params.Add("dir", string(opts.Direction))
	}
	if opts.Limit != 0 {
		params.Add("limit", fmt.Sprintf("%d", opts.Limit))
	}

	// TODO: What if we hit limit, we won't have all the messages

	auth := http.AuthKeys(
		http.NewAuthKey("Authorization", sender),
		http.NewAuthKey("Authorization-Channel", channel),
	)

	resp, err := c.get(ctx, path, params, auth)
	if err != nil {
		return nil, 0, err
	}
	if resp == nil {
		return nil, 0, nil
	}

	var out api.MessagesResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, 0, err
	}

	return out.Messages, out.Index, nil
}

// MessageDecrypt decrypts a remote Event from Messages.
func (c *Client) MessageDecrypt(event *events.Event, kr saltpack.Keyring) (*api.Message, error) {
	decrypted, pk, err := saltpack.SigncryptOpen(event.Data, false, kr)
	if err != nil {
		return nil, err
	}
	var message api.Message
	if err := msgpack.Unmarshal(decrypted, &message); err != nil {
		return nil, err
	}
	message.Sender = pk.ID()
	message.RemoteIndex = event.Index
	message.RemoteTimestamp = tsutil.ConvertMillis(event.Timestamp)
	return &message, nil
}
