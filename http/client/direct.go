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
	"github.com/keys-pub/keys/http"
	"github.com/pkg/errors"
)

// DirectMessageSend sends an encrypted message.
func (c *Client) DirectMessageSend(ctx context.Context, message *api.Message, sender *keys.EdX25519Key, recipient keys.ID) error {
	if message.RemoteTimestamp != 0 {
		return errors.Errorf("remote timestamp should be omitted on send")
	}
	if message.RemoteIndex != 0 {
		return errors.Errorf("remote index should be omitted on send")
	}
	if message.Timestamp == 0 {
		return errors.Errorf("message timestamp is not set")
	}

	encrypted, err := message.Encrypt(sender, recipient)
	if err != nil {
		return err
	}

	path := dstore.Path("dm", recipient, sender.ID())
	vals := url.Values{}
	// vals.Set("expire", expire.String())
	if _, err := c.post(ctx, path, vals, bytes.NewReader(encrypted), http.ContentHash(encrypted), sender); err != nil {
		return err
	}
	return nil
}

// DirectMessages returns encrypted messages (as event.Event) and current index
// from a previous index.
// If truncated, there are more results if you call again with the new index.
// To decrypt to api.Message, use DecryptMessage.
func (c *Client) DirectMessages(ctx context.Context, key *keys.EdX25519Key, opts *MessagesOpts) (*Messages, error) {
	if opts == nil {
		opts = &MessagesOpts{}
	}

	path := dstore.Path("dm", key.ID())
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

	resp, err := c.get(ctx, path, params, key)
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
