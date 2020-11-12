package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// MessagePrepare (RPC) prepares to create a message, the response can be used to show a pending message.
func (s *service) MessagePrepare(ctx context.Context, req *MessagePrepareRequest) (*MessagePrepareResponse, error) {
	message, prepareErr := s.messagePrepare(ctx, req.Sender, req.Channel, req.Text)
	if prepareErr != nil {
		return nil, prepareErr
	}
	return &MessagePrepareResponse{
		Message: message,
	}, nil
}

// MessageCreate (RPC) creates a message for a recipient.
func (s *service) MessageCreate(ctx context.Context, req *MessageCreateRequest) (*MessageCreateResponse, error) {
	message, createErr := s.messageCreate(ctx, req.Sender, req.Channel, req.Text)
	if createErr != nil {
		return nil, createErr
	}
	return &MessageCreateResponse{
		Message: message,
	}, nil
}

// Messages (RPC) lists messages.
func (s *service) Messages(ctx context.Context, req *MessagesRequest) (*MessagesResponse, error) {
	kid, err := keys.ParseID(req.Channel)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid channel")
	}
	channel, err := s.edx25519Key(kid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get channel key")
	}
	mid, err := s.lookup(ctx, req.Member, nil)
	if err != nil {
		return nil, err
	}
	member, err := s.edx25519Key(mid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get sender key")
	}

	if err := s.pullMessages(ctx, channel, member); err != nil {
		return nil, err
	}

	messages, err := s.messages(ctx, channel)
	if err != nil {
		return nil, err
	}
	return &MessagesResponse{
		Messages: messages,
	}, nil
}

// messagePrepare returns a Message for an in progress display. The client
// should then use messageCreate to save the message. This needs to be fast, so
// the client can show the a pending message right away. Preparing before create
// is optional.
func (s *service) messagePrepare(ctx context.Context, sender string, channel string, text string) (*Message, error) {
	if sender == "" {
		return nil, errors.Errorf("no sender specified")
	}
	if channel == "" {
		return nil, errors.Errorf("no channel specified")
	}

	sid, err := s.lookup(ctx, sender, nil)
	if err != nil {
		return nil, err
	}
	senderKey, err := s.key(ctx, sid)
	if err != nil {
		return nil, err
	}

	if channel == "" {
		return nil, errors.Errorf("no channel specified")
	}

	id := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
	message := &Message{
		ID:     id,
		Sender: senderKey,
		Content: &Content{
			Data: []byte(text),
			Type: UTF8Content,
		},
		CreatedAt: tsutil.Millis(s.clock.Now()),
	}

	return message, nil
}

func (s *service) messageCreate(ctx context.Context, sender string, channel string, text string) (*Message, error) {
	if sender == "" {
		return nil, errors.Errorf("no sender specified")
	}
	if channel == "" {
		return nil, errors.Errorf("no channel specified")
	}

	sid, err := s.lookup(ctx, sender, nil)
	if err != nil {
		return nil, err
	}
	senderKey, err := s.edx25519Key(sid)
	if err != nil {
		return nil, err
	}

	cid, err := keys.ParseID(channel)
	if err != nil {
		return nil, err
	}
	channelKey, err := s.edx25519Key(cid)
	if err != nil {
		return nil, err
	}

	// TODO: ID from prepare?
	// TODO: Prev
	id := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
	msg := &api.Message{
		ID: id,
		Content: &api.Content{
			Data: []byte(text),
			Type: api.UTF8Content,
		},
		Sender:    sid,
		CreatedAt: s.clock.Now(),
	}

	if err := s.client.MessageSend(ctx, msg, senderKey, channelKey); err != nil {
		return nil, err
	}
	// TODO: Sync to local

	return s.messageToRPC(ctx, msg)
}

func (s *service) message(ctx context.Context, path string) (*Message, error) {
	doc, err := s.db.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	var event events.Event
	if err := doc.To(&event); err != nil {
		return nil, err
	}

	msg, err := client.DecryptMessage(&event, s.vault)
	if err != nil {
		return nil, err
	}

	return s.messageToRPC(ctx, msg)
}

func (s *service) messageToRPC(ctx context.Context, msg *api.Message) (*Message, error) {
	sender, err := s.key(ctx, msg.Sender)
	if err != nil {
		return nil, err
	}
	return &Message{
		ID:        msg.ID,
		Content:   contentToRPC(msg.Content),
		Sender:    sender,
		CreatedAt: tsutil.Millis(msg.CreatedAt),
	}, nil
}

func contentToRPC(content *api.Content) *Content {
	if content == nil {
		return nil
	}
	return &Content{
		Data: content.Data,
		Type: contentTypeToRPC(content.Type),
	}
}

func contentTypeToRPC(ct api.ContentType) ContentType {
	switch ct {
	case api.UTF8Content:
		return UTF8Content
	default:
		return BinaryContent
	}
}

func (s *service) messages(ctx context.Context, channel *keys.EdX25519Key) ([]*Message, error) {
	path := dstore.Path("channel", channel.ID(), "msgs")
	iter, iterErr := s.db.DocumentIterator(ctx, path, dstore.NoData())
	if iterErr != nil {
		return nil, iterErr
	}
	defer iter.Release()
	messages := make([]*Message, 0, 100)
	for {
		e, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if e == nil {
			break
		}
		logger.Debugf("Message %s", e.Path)
		message, err := s.message(ctx, e.Path)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt < messages[j].CreatedAt
	})

	return messages, nil
}

type channelPullState struct {
	Index int64
}

func (s *service) pullMessages(ctx context.Context, channel *keys.EdX25519Key, member *keys.EdX25519Key) error {
	logger.Infof("Pull messages (%s)...", channel.ID())

	// Get local channel state.
	pullStatePath := dstore.Path("channel", channel.ID(), "pull")
	doc, err := s.db.Get(ctx, pullStatePath)
	if err != nil {
		return err
	}
	var state channelPullState
	if doc != nil {
		if err := doc.To(&state); err != nil {
			logger.Errorf("Invalid channel state: %v", err)
		}
	}

	// Get messages.
	logger.Infof("Pull state: %v", state)
	events, next, err := s.client.Messages(ctx, channel, member, &client.MessagesOpts{Index: state.Index})
	if err != nil {
		return err
	}
	// TODO: If limit hit this doesn't get all messages

	logger.Infof("Received %d messages", len(events))
	for _, event := range events {
		logger.Debugf("Saving message %d", event.Index)
		path := dstore.Path("channel", channel.ID(), "msgs", pad(event.Index))
		if err := s.db.Set(ctx, path, dstore.From(event)); err != nil {
			return err
		}
	}

	// Update state
	state.Index = next
	if err := s.db.Set(ctx, pullStatePath, dstore.From(state)); err != nil {
		return err
	}
	return nil
}

func pad(n int64) string {
	if n > 999999999999999 {
		panic("int too large for padding")
	}
	return fmt.Sprintf("%015d", n)
}
