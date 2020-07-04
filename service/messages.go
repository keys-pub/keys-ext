package service

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/docs"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// MessagePrepare (RPC) prepares to create a message, the response can be used to show a pending message.
func (s *service) MessagePrepare(ctx context.Context, req *MessagePrepareRequest) (*MessagePrepareResponse, error) {
	message, prepareErr := s.messagePrepare(ctx, req.Sender, req.Recipient, req.Text)
	if prepareErr != nil {
		return nil, prepareErr
	}
	return &MessagePrepareResponse{
		Message: message,
	}, nil
}

// MessageCreate (RPC) creates a message for a recipient.
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

	if err := s.pullMessages(ctx, key, recipient); err != nil {
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

	id := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
	message := &Message{
		ID: id,
		Content: &Content{
			Data: []byte(text),
			Type: UTF8Content,
		},
	}

	if err := s.fillMessage(ctx, message, time.Now(), key.ID()); err != nil {
		return nil, err
	}
	return message, nil
}

func (s *service) messageCreate(ctx context.Context, sender string, recipient string, text string) (*Message, error) {
	if recipient == "" {
		return nil, errors.Errorf("no recipient specified")
	}

	key, err := s.parseIdentityForEdX25519Key(ctx, sender)
	if err != nil {
		return nil, err
	}

	rid, err := keys.ParseID(recipient)
	if err != nil {
		return nil, err
	}

	// TODO: ID from prepare?
	id := encoding.MustEncode(keys.RandBytes(32), encoding.Base62)
	message := &Message{
		ID: id,
		Content: &Content{
			Data: []byte(text),
			Type: UTF8Content,
		},
	}
	b, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	event := client.NewEvent(docs.Path("msgs", id), b, nil)
	if err := s.remote.MessageSend(ctx, key, rid, event); err != nil { //, time.Hour*24)
		return nil, err
	}
	// TODO: Sync to local

	return message, nil
}

func (s *service) fillMessage(ctx context.Context, message *Message, t time.Time, sender keys.ID) error {
	key, err := s.keyIDToRPC(ctx, sender)
	if err != nil {
		return err
	}

	message.Sender = key
	message.CreatedAt = int64(tsutil.Millis(t))
	message.TimeDisplay = timeDisplay(t)
	message.DateDisplay = dateDisplay(t)
	return nil
}

func (s *service) message(ctx context.Context, path string) (*Message, error) {
	doc, err := s.db.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	decrypted, sender, err := saltpack.SigncryptOpen(doc.Data, s.vault)
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
	if err := s.fillMessage(ctx, &message, createdAt, kid); err != nil {
		return nil, err
	}
	return &message, nil
}

func (s *service) messages(ctx context.Context, key *keys.EdX25519Key, recipient keys.ID) ([]*Message, error) {
	addr, err := keys.NewAddress(key.ID(), recipient)
	if err != nil {
		return nil, err
	}
	path := docs.Path("msgs", addr)
	iter, iterErr := s.db.DocumentIterator(ctx, path, docs.NoData())
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

func (s *service) pullMessages(ctx context.Context, key *keys.EdX25519Key, recipient keys.ID) error {
	logger.Infof("Pull messages...")
	return errors.Errorf("not implemented")
}
