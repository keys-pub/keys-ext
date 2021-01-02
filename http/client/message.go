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

// Messages response.
type Messages struct {
	Events    []*events.Event
	Index     int64
	Truncated bool
}

// Decrypt messages.
func (m Messages) Decrypt(key *keys.EdX25519Key) ([]*api.Message, error) {
	msgs := make([]*api.Message, 0, len(m.Events))
	for _, event := range m.Events {
		msg, err := api.DecryptMessageFromEvent(event, key)
		if err != nil {
			// TODO: Skip invalid messages
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

// Messages returns encrypted messages (as event.Event) and current index from a
// previous index.
// If truncated, there are more results if you call again with the new index.
// To decrypt to api.Message, use DecryptMessage.
func (c *Client) Messages(ctx context.Context, channel *keys.EdX25519Key, opts *MessagesOpts) (*Messages, error) {
	path := dstore.Path("channel", channel.ID(), "msgs")
	return c.messages(ctx, path, channel, opts)
}

func (c *Client) messages(ctx context.Context, path string, key *keys.EdX25519Key, opts *MessagesOpts) (*Messages, error) {
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

	var out api.MessagesResponse
	if err := json.Unmarshal(resp.Data, &out); err != nil {
		return nil, err
	}

	return &Messages{
		Events:    out.Messages,
		Index:     out.Index,
		Truncated: out.Truncated,
	}, nil
}
