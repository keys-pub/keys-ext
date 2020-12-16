package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// MessagePrepare returns a Message for an in progress display. The client
// should then use messageCreate to save the message. Prepare needs to be fast,
// so the client can show the a pending message right away.
// Preparing before create is optional.
func (s *service) MessagePrepare(ctx context.Context, req *MessagePrepareRequest) (*MessagePrepareResponse, error) {
	if req.Sender == "" {
		return nil, errors.Errorf("no sender specified")
	}
	if req.Channel == "" {
		return nil, errors.Errorf("no channel specified")
	}

	sender, err := s.lookup(ctx, req.Sender, nil)
	if err != nil {
		return nil, err
	}
	senderKey, err := s.key(ctx, sender)
	if err != nil {
		return nil, err
	}

	text := processText(req.Text)

	id := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
	message := &Message{
		ID:        id,
		Sender:    senderKey,
		Text:      []string{text},
		Status:    MessagePending,
		CreatedAt: tsutil.Millis(s.clock.Now()),
	}

	return &MessagePrepareResponse{
		Message: message,
	}, nil
}

func processText(s string) string {
	return strings.TrimSpace(s)
}

// MessageCreate (RPC) creates a message for a recipient.
func (s *service) MessageCreate(ctx context.Context, req *MessageCreateRequest) (*MessageCreateResponse, error) {
	if req.Sender == "" {
		return nil, errors.Errorf("no sender specified")
	}
	if req.Channel == "" {
		return nil, errors.Errorf("no channel specified")
	}

	text := processText(req.Text)
	if strings.HasPrefix(text, "/") {
		msg, err := s.channelCommand(ctx, text, req.Sender, req.Channel)
		if err != nil {
			return nil, err
		}
		return &MessageCreateResponse{Message: msg}, nil
	}

	sender, err := s.lookup(ctx, req.Sender, nil)
	if err != nil {
		return nil, err
	}
	senderKey, err := s.edx25519Key(sender)
	if err != nil {
		return nil, err
	}

	channel, err := keys.ParseID(req.Channel)
	if err != nil {
		return nil, err
	}
	channelKey, err := s.edx25519Key(channel)
	if err != nil {
		return nil, err
	}

	// TODO: Prev
	id := req.ID
	if id == "" {
		id = encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
	}
	msg := &api.Message{
		ID:        id,
		Text:      text,
		Sender:    sender,
		Timestamp: s.clock.NowMillis(),
	}

	if err := s.client.MessageSend(ctx, msg, senderKey, channelKey); err != nil {
		return nil, err
	}

	out, err := s.messageToRPC(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &MessageCreateResponse{
		Message: out,
	}, nil
}

// Messages (RPC) lists messages.
func (s *service) Messages(ctx context.Context, req *MessagesRequest) (*MessagesResponse, error) {
	channel, err := keys.ParseID(req.Channel)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid channel")
	}
	user, err := s.lookup(ctx, req.User, nil)
	if err != nil {
		return nil, err
	}

	if req.Update {
		if err := s.pullMessages(ctx, channel, user); err != nil {
			return nil, err
		}
	}

	messages, err := s.messages(ctx, channel)
	if err != nil {
		return nil, err
	}

	return &MessagesResponse{
		Messages: messages,
	}, nil
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

	msg, err := api.DecryptMessage(&event, s.vault)
	if err != nil {
		return nil, err
	}

	return s.messageToRPC(ctx, msg)
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

func (s *service) pullMessages(ctx context.Context, channel keys.ID, user keys.ID) error {
	channelKey, err := s.edx25519Key(channel)
	if err != nil {
		return err
	}
	userKey, err := s.edx25519Key(user)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			truncated, err := s.pullMessagesNext(ctx, channelKey, userKey)
			if err != nil {
				return err
			}
			if !truncated {
				return nil
			}
		}
	}
}

func (s *service) pullMessagesNext(ctx context.Context, channelKey *keys.EdX25519Key, userKey *keys.EdX25519Key) (bool, error) {
	logger.Infof("Pull messages (%s)...", channelKey.ID())

	pullState, err := s.channelPullState(ctx, channelKey.ID())
	if err != nil {
		return false, err
	}

	// Get messages.
	logger.Infof("Pull state: %v", pullState)
	msgs, err := s.client.Messages(ctx, channelKey, userKey, &client.MessagesOpts{Index: pullState.Index})
	if err != nil {
		return false, err
	}

	logger.Infof("Received %d messages", len(msgs.Messages))

	info := &channelInfo{ID: channelKey.ID()}

	for _, event := range msgs.Messages {
		logger.Debugf("Saving message %d", event.Index)
		path := dstore.Path("messages", channelKey.ID(), pad(event.Index))
		if err := s.db.Set(ctx, path, dstore.From(event)); err != nil {
			return false, err
		}

		// TODO: Gracefully handle message errors (continue?)

		// Decrypt message temporarily to update channel state.
		msg, err := api.DecryptMessage(event, s.vault)
		if err != nil {
			return false, err
		}

		// Update snippet
		rmsg, err := s.messageToRPC(ctx, msg)
		if err != nil {
			return false, err
		}
		if len(rmsg.Text) > 0 {
			info.Snippet = rmsg.Text[len(rmsg.Text)-1]
		}

		// Update channel info
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
	info.Index = msgs.Index

	// Save channel state.
	if err := s.db.Set(ctx, dstore.Path("channel", channelKey.ID(), "info"), dstore.From(info), dstore.MergeAll()); err != nil {
		return false, err
	}

	// Update pull state.
	pullState.Index = msgs.Index
	if err := s.db.Set(ctx, dstore.Path("channel", channelKey.ID(), "pull"), dstore.From(pullState)); err != nil {
		return false, err
	}

	return msgs.Truncated, nil
}

func (s *service) channelPullState(ctx context.Context, channel keys.ID) (*channelPullState, error) {
	pullStatePath := dstore.Path("channel", channel, "pull")
	var pullState channelPullState
	if _, err := s.db.Load(ctx, pullStatePath, &pullState); err != nil {
		return nil, err
	}
	return &pullState, nil
}

func pad(n int64) string {
	if n > 999999999999999 {
		panic("int too large for padding")
	}
	return fmt.Sprintf("%015d", n)
}

func (s *service) messageToRPC(ctx context.Context, msg *api.Message) (*Message, error) {
	if msg == nil {
		return nil, nil
	}
	if msg.Sender == "" {
		return nil, errors.Errorf("no sender")
	}

	sender, err := s.key(ctx, msg.Sender)
	if err != nil {
		return nil, err
	}
	text, err := s.messageText(ctx, msg, sender)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:        msg.ID,
		Text:      text,
		Sender:    sender,
		CreatedAt: msg.Timestamp,
	}, nil
}

func (s *service) messageText(ctx context.Context, msg *api.Message, sender *Key) ([]string, error) {
	texts := []string{}
	if msg.Text != "" {
		texts = append(texts, msg.Text)
	}

	// Info
	if msg.ChannelInfo != nil && msg.ChannelInfo.Name != "" {
		texts = append(texts, fmt.Sprintf("%s set the channel name to %s", sender.userName(), msg.ChannelInfo.Name))
	}
	if msg.ChannelInfo != nil && msg.ChannelInfo.Description != "" {
		texts = append(texts, fmt.Sprintf("%s set the channel description to %s", sender.userName(), msg.ChannelInfo.Description))
	}

	// Notifications
	if msg.ChannelInvites != nil {
		uks, err := s.resolveKeys(ctx, msg.ChannelInvites.Users)
		if err != nil {
			return nil, err
		}
		texts = append(texts, fmt.Sprintf("%s invited %s", sender.userName(), keyUserNames(uks)))
	}
	if msg.ChannelUninvites != nil {
		uks, err := s.resolveKeys(ctx, msg.ChannelUninvites.Users)
		if err != nil {
			return nil, err
		}
		texts = append(texts, fmt.Sprintf("%s uninvited %s", sender.userName(), keyUserNames(uks)))
	}
	if msg.ChannelJoin != nil {
		texts = append(texts, fmt.Sprintf("%s joined", sender.userName()))
	}
	if msg.ChannelLeave != nil {
		texts = append(texts, fmt.Sprintf("%s left", sender.userName()))
	}

	return texts, nil
}
