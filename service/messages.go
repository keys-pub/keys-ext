package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keysd/http/client"
	"github.com/pkg/errors"
)

// MessagePrepare (RPC) prepares to create a message, the response can be used to show a pending message
func (s *service) MessagePrepare(ctx context.Context, req *MessagePrepareRequest) (*MessagePrepareResponse, error) {
	message, prepareErr := s.messagePrepare(ctx, req.Sender, req.Recipient, req.Text)
	if prepareErr != nil {
		return nil, prepareErr
	}
	return &MessagePrepareResponse{
		Message: message,
	}, nil
}

// MessageCreate (RPC) creates a message for a recipient
func (s *service) MessageCreate(ctx context.Context, req *MessageCreateRequest) (*MessageCreateResponse, error) {
	message, createErr := s.messageCreate(ctx, req.Sender, req.Recipient, req.Text)
	if createErr != nil {
		return nil, createErr
	}
	return &MessageCreateResponse{
		Message: message,
	}, nil
}

// Messages (RPC) lists messages.
func (s *service) Messages(ctx context.Context, req *MessagesRequest) (*MessagesResponse, error) {
	if req.Sender == "" {
		return nil, errors.Errorf("no kid specified")
	}
	key, err := s.parseSignKey(req.Sender, true)
	if err != nil {
		return nil, err
	}
	if req.Recipient == "" {
		return nil, errors.Errorf("no recipient")
	}
	recipient, err := keys.ParseID(req.Recipient)
	if err != nil {
		return nil, err
	}

	if err := s.pullMessages(ctx, key.ID(), recipient); err != nil {
		return nil, err
	}

	messages, err := s.messages(ctx, key, recipient)
	if err != nil {
		return nil, err
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
func (s *service) messagePrepare(ctx context.Context, sender string, recipient string, text string) (*Message, error) {
	if sender == "" {
		return nil, errors.Errorf("no sender specified")
	}
	key, err := s.parseSignKey(sender, true)
	if err != nil {
		return nil, err
	}
	if recipient == "" {
		return nil, errors.Errorf("no recipient specified")
	}
	_, err = keys.ParseID(recipient)
	if err != nil {
		return nil, err
	}

	message := &Message{
		Content: &MessageContent{
			Text: text,
		},
	}

	s.fillMessage(ctx, message, time.Now(), key.ID(), "")
	return message, nil
}

func (s *service) messageCreate(ctx context.Context, sender string, recipient string, text string) (*Message, error) {
	if recipient == "" {
		return nil, errors.Errorf("no recipient specified")
	}

	senderKey, err := s.parseSignKey(sender, true)
	if err != nil {
		return nil, err
	}

	rid, err := keys.ParseID(recipient)
	if err != nil {
		return nil, err
	}

	message := &Message{
		Content: &MessageContent{
			Text: text,
		},
	}
	b, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	sent, err := s.remote.SendMessage(senderKey, rid, b, nil)
	if err != nil {
		return nil, err
	}

	message.ID = sent.ID

	// TODO: Sync to local

	return message, nil
}

func (s *service) fillMessage(ctx context.Context, message *Message, t time.Time, sender keys.ID, path string) {
	result, err := s.users.Get(ctx, sender)
	if err != nil {
		logger.Errorf("Failed to load sigchain: %s", err)
	}

	message.User = userResultToRPC(result)
	message.Sender = sender.String()
	message.CreatedAt = int64(keys.TimeToMillis(t))
	message.TimeDisplay = timeDisplay(t)
	message.DateDisplay = dateDisplay(t)
	message.Path = path
}

func (s *service) message(ctx context.Context, path string) (*Message, error) {
	doc, err := s.db.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	sp := saltpack.NewSaltpack(s.ks)
	decrypted, sender, err := sp.SigncryptOpen(doc.Data)
	if err != nil {
		return nil, err
	}

	var message Message
	if err := json.Unmarshal(decrypted, &message); err != nil {
		return nil, err
	}
	createdAt := doc.CreatedAt
	var kid keys.ID
	if sender != nil {
		kid = sender.ID()
	}
	s.fillMessage(ctx, &message, createdAt, kid, path)
	return &message, nil
}

func (s *service) messages(ctx context.Context, key *keys.EdX25519Key, recipient keys.ID) ([]*Message, error) {
	path := fmt.Sprintf("messages-%s-%s", key.ID(), recipient)
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

func (s *service) pullMessages(ctx context.Context, kid keys.ID, recipient keys.ID) error {
	key, err := s.ks.EdX25519Key(kid)
	if err != nil {
		return err
	}
	if key == nil {
		return keys.NewErrNotFound(kid.String())
	}
	logger.Infof("Pull messages...")
	versionPath := keys.Path("versions", fmt.Sprintf("messages-%s", kid))
	e, err := s.db.Get(ctx, versionPath)
	if err != nil {
		return err
	}
	version := ""
	if e != nil {
		version = string(e.Data)
	}
	msgs, version, err := s.remote.Messages(key, recipient, &client.MessagesOpts{Version: version})
	if err != nil {
		return err
	}
	// TODO: If limit hit this doesn't get all messages

	logger.Infof("Received %d messages", len(msgs))
	for _, msg := range msgs {
		ts := 9223372036854775807 - keys.TimeToMillis(msg.CreatedAt)
		pathKey := fmt.Sprintf("messages-%s-%s", key.ID(), recipient)
		pathVal := fmt.Sprintf("%d-%s", ts, msg.ID)
		path := keys.Path(pathKey, pathVal)
		if err := s.db.Set(ctx, path, msg.Data); err != nil {
			return err
		}
	}
	if err := s.db.Set(ctx, versionPath, []byte(version)); err != nil {
		return err
	}
	return nil
}
