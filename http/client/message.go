package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/dstore/events"
)

// MessageSend sends an encrypted message to a channel.
func (c *Client) MessageSend(ctx context.Context, message *api.Message, sender *keys.EdX25519Key, channel *keys.EdX25519Key) error {
	encrypted, err := message.Encrypt(sender, channel.ID())
	if err != nil {
		return err
	}

	path := dstore.Path("channel", channel.ID(), "msgs")
	vals := url.Values{}
	// vals.Set("expire", expire.String())
	if _, err := c.req(ctx, request{Method: "POST", Path: path, Params: vals, Body: encrypted, Key: channel}); err != nil {
		return err
	}
	return nil
}

// MessagesOpts options for Messages.
type MessagesOpts struct {
	// Index to list to/from
	Index int64
	// Order ascending or descending
	Order events.Direction
	// Limit by
	Limit int
}

// Messages returns encrypted messages (as event.Event) and current index from a
// previous index.
// If truncated, there are more results if you call again with the new index.
// To decrypt to api.Event, use api.DecryptMessageFromEvent.
func (c *Client) Messages(ctx context.Context, channel *keys.EdX25519Key, opts *MessagesOpts) (*api.Events, error) {
	path := dstore.Path("channel", channel.ID(), "msgs")
	return c.events(ctx, path, channel, opts)
}

func (c *Client) events(ctx context.Context, path string, key *keys.EdX25519Key, opts *MessagesOpts) (*api.Events, error) {
	if opts == nil {
		opts = &MessagesOpts{}
	}

	params := url.Values{}
	if opts.Index != 0 {
		params.Add("idx", strconv.FormatInt(opts.Index, 10))
	}
	if opts.Order != "" {
		params.Add("order", string(opts.Order))
	}
	if opts.Limit != 0 {
		params.Add("limit", fmt.Sprintf("%d", opts.Limit))
	}

	resp, err := c.req(ctx, request{
		Method: "GET",
		Path:   path,
		Params: params,
		Key:    key,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	var out api.Events
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}

	return &out, nil
}
