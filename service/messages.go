package service

import (
	"context"
	"fmt"

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
	channel, err := keys.ParseID(req.Channel)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid channel")
	}
	member, err := s.lookup(ctx, req.Member, nil)
	if err != nil {
		return nil, err
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
		ID:        id,
		Sender:    senderKey,
		Text:      []string{text},
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
		ID:        id,
		Text:      text,
		Sender:    sid,
		Timestamp: s.clock.NowMillis(),
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

func keyUserName(key *Key) string {
	if key.User != nil && key.User.Name != "" {
		return key.User.Name
	}
	return key.ID
}

func (s *service) messageToRPC(ctx context.Context, msg *api.Message) (*Message, error) {
	sender, err := s.key(ctx, msg.Sender)
	if err != nil {
		return nil, err
	}
	texts := []string{}
	if msg.Text != "" {
		texts = append(texts, msg.Text)
	}
	if msg.ChannelInfo != nil && msg.ChannelInfo.Name != "" {
		texts = append(texts, fmt.Sprintf("%s set the name to %q", keyUserName(sender), msg.ChannelInfo.Name))
	}
	if msg.ChannelInfo != nil && msg.ChannelInfo.Description != "" {
		texts = append(texts, fmt.Sprintf("%s set the description to %q", keyUserName(sender), msg.ChannelInfo.Description))
	}

	return &Message{
		ID:        msg.ID,
		Text:      texts,
		Sender:    sender,
		CreatedAt: msg.Timestamp,
	}, nil
}

func (s *service) messages(ctx context.Context, channel keys.ID) ([]*Message, error) {
	path := dstore.Path("messages", channel.ID())
	iter, err := s.db.DocumentIterator(ctx, path, dstore.NoData())
	if err != nil {
		return nil, err
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
	return messages, nil
}

type channelPullState struct {
	Index int64 `json:"idx"`
}

func (s *service) pullMessages(ctx context.Context, channel keys.ID, member keys.ID) error {
	channelKey, err := s.edx25519Key(channel)
	if err != nil {
		return err
	}
	memberKey, err := s.edx25519Key(member)
	if err != nil {
		return err
	}

	logger.Infof("Pull messages (%s)...", channel)

	// Keep pulling messages til we receive none.
	for {
		count, err := s.pullMessagesNext(ctx, channelKey, memberKey)
		if err != nil {
			return err
		}
		if count == 0 {
			return nil
		}
	}
}

func (s *service) channelPullState(ctx context.Context, channel keys.ID) (*channelPullState, error) {
	pullStatePath := dstore.Path("channel", channel, "pull")
	var pullState channelPullState
	if _, err := s.db.Load(ctx, pullStatePath, &pullState); err != nil {
		return nil, err
	}
	return &pullState, nil
}

func (s *service) pullMessagesNext(ctx context.Context, channel *keys.EdX25519Key, member *keys.EdX25519Key) (int, error) {
	pullState, err := s.channelPullState(ctx, channel.ID())
	if err != nil {
		return 0, err
	}

	// Get messages.
	logger.Infof("Pull state: %v", pullState)
	events, next, err := s.client.Messages(ctx, channel, member, &client.MessagesOpts{Index: pullState.Index})
	if err != nil {
		return 0, err
	}

	logger.Infof("Received %d messages", len(events))

	info := &channelInfo{ID: channel.ID()}

	for _, event := range events {
		logger.Debugf("Saving message %d", event.Index)
		path := dstore.Path("messages", channel.ID(), pad(event.Index))
		if err := s.db.Set(ctx, path, dstore.From(event)); err != nil {
			return 0, err
		}

		// Decrypt message temporarily to update channel state.
		msg, err := client.DecryptMessage(event, s.vault)
		if err != nil {
			return 0, err
		}
		if msg.Text != "" {
			info.Snippet = msg.Text
		}
		if msg.ChannelInfo != nil {
			if msg.ChannelInfo.Name != "" {
				info.Name = msg.ChannelInfo.Name
			}
			if msg.ChannelInfo.Description != "" {
				info.Description = msg.ChannelInfo.Description
			}
		}
		info.Timestamp = msg.Timestamp
		info.RemoteTimestamp = msg.RemoteTimestamp
	}

	// Save channel state.
	if err := s.db.Set(ctx, dstore.Path("channel", channel.ID(), "info"), dstore.From(info), dstore.MergeAll()); err != nil {
		return 0, err
	}

	// Update pull state
	pullState.Index = next
	if err := s.db.Set(ctx, dstore.Path("channel", channel.ID(), "pull"), dstore.From(pullState)); err != nil {
		return 0, err
	}

	return len(events), nil
}

func pad(n int64) string {
	if n > 999999999999999 {
		panic("int too large for padding")
	}
	return fmt.Sprintf("%015d", n)
}
