package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keysd/http/api"
	"github.com/pkg/errors"
)

// Message from server.
type Message struct {
	ID   string
	Data []byte

	CreatedAt time.Time
	UpdatedAt time.Time
}

// MessageOpts are options for SendMessage.
type MessageOpts struct {
	// Channel to post on.
	Channel string
}

// SendMessage posts an encrypted message.
func (c *Client) SendMessage(ctx context.Context, sender *keys.EdX25519Key, recipient keys.ID, b []byte, opts *MessageOpts) (*Message, error) {
	sp := saltpack.NewSaltpack(c.ks)
	encrypted, err := sp.Signcrypt(b, sender, recipient, sender.ID())
	if err != nil {
		return nil, err
	}
	return c.postMessage(ctx, sender, recipient, encrypted, opts)
}

func (c *Client) postMessage(ctx context.Context, sender *keys.EdX25519Key, recipient keys.ID, b []byte, opts *MessageOpts) (*Message, error) {
	if opts == nil {
		opts = &MessageOpts{}
	}
	path := keys.Path("messages", sender.ID(), recipient)
	vals := url.Values{}
	if opts.Channel != "" {
		vals.Add("channel", opts.Channel)
	}
	doc, err := c.postDocument(ctx, path, vals, sender, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.Errorf("failed to post message: no response")
	}

	var msg api.MessageResponse
	if err := json.Unmarshal(doc.Data, &msg); err != nil {
		return nil, err
	}
	// TODO: CreatedAt, UpdatedAt
	return &Message{
		ID:   msg.ID,
		Data: b,
	}, nil
}

// MessagesOpts options for Messages.
type MessagesOpts struct {
	// Version to list to/from
	Version string
	// Direction ascending or descending
	Direction keys.Direction
	// Channel to filter by
	Channel string
	// Limit by
	Limit int
}

// Messages returns encrypted messages.
// To decrypt a message, use Client#DecryptMessage.
func (c *Client) Messages(ctx context.Context, key *keys.EdX25519Key, from keys.ID, opts *MessagesOpts) ([]*Message, string, error) {
	path := keys.Path("messages", key.ID(), from)
	if opts == nil {
		opts = &MessagesOpts{}
	}

	params := url.Values{}
	params.Add("include", "md")
	if opts.Version != "" {
		params.Add("version", opts.Version)
	}
	if opts.Direction != "" {
		params.Add("direction", string(opts.Direction))
	}
	if opts.Channel != "" {
		params.Add("channel", opts.Channel)
	}
	if opts.Limit != 0 {
		params.Add("limit", fmt.Sprintf("%d", opts.Limit))
	}

	// TODO: What if we hit limit, we won't have all the messages

	doc, err := c.getDocument(ctx, path, params, key)
	if err != nil {
		return nil, "", err
	}
	if doc == nil {
		return nil, "", nil
	}

	var resp api.MessagesResponse
	if err := json.Unmarshal(doc.Data, &resp); err != nil {
		return nil, "", err
	}

	msgs := make([]*Message, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		msgs = append(msgs, &Message{
			ID:        msg.ID,
			Data:      msg.Data,
			CreatedAt: resp.MetadataFor(msg).CreatedAt,
			UpdatedAt: resp.MetadataFor(msg).UpdatedAt,
		})
	}

	return msgs, resp.Version, nil
}

func (c *Client) DecryptMessage(key *keys.EdX25519Key, msg *Message) ([]byte, keys.ID, error) {
	sp := saltpack.NewSaltpack(c.ks)
	decrypted, pk, err := sp.SigncryptOpen(msg.Data)
	if err != nil {
		return nil, "", err
	}
	return decrypted, pk.ID(), nil
}
