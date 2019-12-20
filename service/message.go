package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// MessagePrepare (RPC) prepares to create a message, the response can be used to show a pending message
func (s *service) MessagePrepare(ctx context.Context, req *MessagePrepareRequest) (*MessagePrepareResponse, error) {
	message, prepareErr := s.messagePrepare(ctx, req.Sender, req.KID, req.Text)
	if prepareErr != nil {
		return nil, prepareErr
	}
	return &MessagePrepareResponse{
		Message: message,
	}, nil
}

// MessageCreate (RPC) creates a message for a recipient
func (s *service) MessageCreate(ctx context.Context, req *MessageCreateRequest) (*MessageCreateResponse, error) {
	message, createErr := s.messageCreate(ctx, req.Sender, req.KID, req.ID, req.Text)
	if createErr != nil {
		return nil, createErr
	}
	return &MessageCreateResponse{
		Message: message,
	}, nil
}

// Messages (RPC) lists messages in a group
func (s *service) Messages(ctx context.Context, req *MessagesRequest) (*MessagesResponse, error) {
	if req.KID == "" {
		return nil, errors.Errorf("no kid specified")
	}
	key, err := s.parseKey(req.KID)
	if err != nil {
		return nil, err
	}

	if err := s.pullMessages(ctx, key.ID()); err != nil {
		return nil, err
	}

	messages, messagesErr := s.messages(ctx, key)
	if messagesErr != nil {
		return nil, messagesErr
	}
	return &MessagesResponse{
		Messages: messages,
	}, nil
}

// Inbox (RPC)
func (s *service) Inbox(ctx context.Context, req *InboxRequest) (*InboxResponse, error) {
	return nil, errors.Errorf("not implemented")
}

// messagePrepare returns a Message for an in progress display. The client
// should then use messageCreate to save the message. This needs to be fast, so
// the client can show the a pending message right away. Preparing before create
// is optional.
func (s *service) messagePrepare(ctx context.Context, sender string, kid string, text string) (*Message, error) {
	if kid == "" {
		return nil, errors.Errorf("no kid specified")
	}
	key, err := s.parseKeyOrCurrent(sender)
	if err != nil {
		return nil, err
	}
	_, err = s.parseKey(kid)
	if err != nil {
		return nil, err
	}

	message := &Message{
		ID: keys.RandID().String(),
		Content: &MessageContent{
			Text: text,
		},
	}

	s.fillMessage(ctx, message, time.Now(), key.ID(), "")
	return message, nil
}

func (s *service) messageCreate(ctx context.Context, sender string, kid string, rawid string, text string) (*Message, error) {
	if kid == "" {
		return nil, errors.Errorf("no kid specified")
	}
	var id keys.ID
	if rawid == "" {
		id = keys.RandID()
	} else {
		i, err := keys.ParseID(rawid)
		if err != nil {
			return nil, err
		}
		id = i
	}

	senderKey, err := s.parseKeyOrCurrent(sender)
	if err != nil {
		return nil, err
	}

	// key, err := s.parseShareKey(kid, senderKey)
	// if err != nil {
	// 	return nil, err
	// }
	key, err := s.parseKey(kid)
	if err != nil {
		return nil, err
	}

	message := &Message{
		ID: id.String(),
		Content: &MessageContent{
			Text: text,
		},
	}
	b, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	if s.remote == nil {
		return nil, errors.Errorf("no remote set")
	}
	if _, err := s.remote.PutMessage(senderKey, key, id, b); err != nil {
		return nil, err
	}

	// TODO: Sync to local

	return message, nil
}

func (s *service) fillMessage(ctx context.Context, message *Message, t time.Time, sender keys.ID, path string) {
	user, resolveErr := s.findUser(ctx, sender)
	if resolveErr != nil {
		logger.Errorf("Failed to resolve user: %s", resolveErr)
	}
	message.User = userToRPC(user)
	message.Sender = sender.String()
	message.CreatedAt = int64(keys.TimeToMillis(t))
	message.TimeDisplay = timeDisplay(t)
	message.DateDisplay = dateDisplay(t)
	message.Path = path
}

func (s *service) message(ctx context.Context, path string) (*Message, error) {
	opened, err := s.local.Open(ctx, path)
	if err != nil {
		return nil, err
	}
	if opened == nil {
		return nil, nil
	}
	var message Message
	if err := json.Unmarshal(opened.Data, &message); err != nil {
		return nil, err
	}
	createdAt := opened.Document.CreatedAt
	s.fillMessage(ctx, &message, createdAt, opened.Signer, path)
	return &message, nil
}

func (s *service) messages(ctx context.Context, key keys.Key) ([]*Message, error) {
	path := fmt.Sprintf("messages-%s", key.ID())
	iter, iterErr := s.db.Documents(ctx, path, &keys.DocumentsOpts{PathOnly: true})
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
		message, messageErr := s.message(ctx, e.Path)
		if messageErr != nil {
			return nil, messageErr
		}
		messages = append(messages, message)
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt < messages[j].CreatedAt
	})

	return messages, nil
}
