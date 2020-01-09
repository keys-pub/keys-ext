package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/saltpack"
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

// Messages (RPC) lists messages.
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
		ID: keys.RandString(32),
		Content: &MessageContent{
			Text: text,
		},
	}

	s.fillMessage(ctx, message, time.Now(), key.ID(), "")
	return message, nil
}

func (s *service) messageCreate(ctx context.Context, sender string, kid string, id string, text string) (*Message, error) {
	if kid == "" {
		return nil, errors.Errorf("no kid specified")
	}
	if id == "" {
		id = keys.RandString(32)
	}

	senderKey, err := s.parseKeyOrCurrent(sender)
	if err != nil {
		return nil, err
	}

	key, err := s.parseKey(kid)
	if err != nil {
		return nil, err
	}

	message := &Message{
		ID: id,
		Content: &MessageContent{
			Text: text,
		},
	}
	b, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	sp := saltpack.NewSaltpack(s.ks)
	encrypted, err := sp.Signcrypt(b, senderKey, key.ID())
	if err != nil {
		return nil, err
	}

	if s.remote == nil {
		return nil, errors.Errorf("no remote set")
	}
	if err := s.remote.PutMessage(key, id, encrypted); err != nil {
		return nil, err
	}

	// TODO: Sync to local

	return message, nil
}

func (s *service) fillMessage(ctx context.Context, message *Message, t time.Time, sender keys.ID, path string) {
	res, err := s.users.Get(ctx, sender)
	if err != nil {
		logger.Errorf("Failed to load sigchain: %s", err)
	}

	message.Users = userResultsToRPC(res)
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
	s.fillMessage(ctx, &message, createdAt, sender, path)
	return &message, nil
}

func (s *service) messages(ctx context.Context, key *keys.SignKey) ([]*Message, error) {
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

func (s *service) pullMessages(ctx context.Context, kid keys.ID) error {
	key, err := s.ks.SignKey(kid)
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
	if s.remote == nil {
		return errors.Errorf("no remote set")
	}
	resp, err := s.remote.Messages(key, version)
	if err != nil {
		return err
	}
	if resp == nil {
		logger.Infof("No messages")
		return nil
	}
	logger.Infof("Received %d messages", len(resp.Messages))
	for _, msg := range resp.Messages {
		md := resp.MetadataFor(msg)
		ts := 9223372036854775807 - keys.TimeToMillis(md.CreatedAt)
		pathKey := fmt.Sprintf("messages-%s", key.ID())
		pathVal := fmt.Sprintf("%d-%s", ts, msg.ID)
		path := keys.Path(pathKey, pathVal)
		if err := s.db.Set(ctx, path, msg.Data); err != nil {
			return err
		}
	}
	if err := s.db.Set(ctx, versionPath, []byte(resp.Version)); err != nil {
		return err
	}
	return nil
}
