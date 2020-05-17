package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
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

// SendMessage posts an encrypted expiring message.
func (c *Client) SendMessage(ctx context.Context, sender *keys.EdX25519Key, recipient keys.ID, b []byte, expire time.Duration) (*api.CreateMessageResponse, error) {
	if expire == time.Duration(0) {
		return nil, errors.Errorf("no expire specified")
	}
	encrypted, err := saltpack.Signcrypt(b, sender, recipient, sender.ID())
	if err != nil {
		return nil, err
	}

	path := ds.Path("msgs", sender.ID(), recipient)
	vals := url.Values{}
	vals.Set("expire", expire.String())
	doc, err := c.postDocument(ctx, path, vals, sender, bytes.NewReader(encrypted))
	if err != nil {
		return nil, err
	}
	var resp api.CreateMessageResponse
	if err := json.Unmarshal(doc.Data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MessagesOpts options for Messages.
type MessagesOpts struct {
	// Version to list to/from
	Version string
	// Direction ascending or descending
	Direction ds.Direction
	// Limit by
	Limit int
}

// Messages returns encrypted messages.
// To decrypt a message, use Client#DecryptMessage.
func (c *Client) Messages(ctx context.Context, key *keys.EdX25519Key, from keys.ID, opts *MessagesOpts) ([]*Message, string, error) {
	path := ds.Path("msgs", key.ID(), from)
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

// DecryptMessage decrypts a message from Messages.
func (c *Client) DecryptMessage(key *keys.EdX25519Key, msg *Message) ([]byte, keys.ID, error) {
	decrypted, pk, err := saltpack.SigncryptOpen(msg.Data, saltpack.NewKeyStore(key))
	if err != nil {
		return nil, "", err
	}
	return decrypted, pk.ID(), nil
}
