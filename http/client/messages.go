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
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/docs/events"
	"github.com/keys-pub/keys/saltpack"
	"github.com/vmihailenco/msgpack/v4"
)

// MessageSend posts an encrypted message.
// TODO: expire time.Duration
func (c *Client) MessageSend(ctx context.Context, sender *keys.EdX25519Key, recipient keys.ID, event *Event) error {
	// if expire == time.Duration(0) {
	// 	return errors.Errorf("no expire specified")
	// }
	b, err := msgpack.Marshal(event)
	if err != nil {
		return err
	}

	encrypted, err := saltpack.Signcrypt(b, sender, recipient, sender.ID())
	if err != nil {
		return err
	}
	contentHash := api.ContentHash(encrypted)

	path := docs.Path("msgs", sender.ID(), recipient)
	vals := url.Values{}
	// vals.Set("expire", expire.String())
	if _, err := c.postDocument(ctx, path, vals, sender, bytes.NewReader(encrypted), contentHash); err != nil {
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
func (c *Client) Messages(ctx context.Context, key *keys.EdX25519Key, from keys.ID, opts *MessagesOpts) ([]*events.Event, int64, error) {
	path := docs.Path("msgs", key.ID(), from)
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

	doc, err := c.getDocument(ctx, path, params, key)
	if err != nil {
		return nil, 0, err
	}
	if doc == nil {
		return nil, 0, nil
	}

	var resp api.MessagesResponse
	if err := json.Unmarshal(doc.Data, &resp); err != nil {
		return nil, 0, err
	}

	return resp.Messages, resp.Index, nil
}

// MessageDecrypt decrypts a remote Event from Messages.
func (c *Client) MessageDecrypt(key *keys.EdX25519Key, revent *events.Event) (*Event, keys.ID, error) {
	decrypted, pk, err := saltpack.SigncryptOpen(revent.Data, saltpack.NewKeyStore(key))
	if err != nil {
		return nil, "", err
	}
	var event Event
	if err := msgpack.Unmarshal(decrypted, &event); err != nil {
		return nil, "", err
	}
	event.Index = revent.Index
	event.Timestamp = revent.Timestamp
	return &event, pk.ID(), nil
}
