package api

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/dstore/events"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

// Encrypt does crypto_box_seal(pk+crypto_box(msgpack(i))).
func Encrypt(i interface{}, sender *keys.EdX25519Key, recipient keys.ID) ([]byte, error) {
	pk := api.NewKey(recipient).AsX25519Public()
	if pk == nil {
		return nil, errors.Errorf("invalid message recipient")
	}
	b, err := msgpack.Marshal(i)
	if err != nil {
		return nil, err
	}
	sk := sender.X25519Key()
	encrypted := keys.BoxSeal(b, pk, sk)
	box := append(sk.Public(), encrypted...)
	anonymized := keys.CryptoBoxSeal(box, pk)
	return anonymized, nil
}

// DecryptMessage decrypts message.
func DecryptMessage(b []byte, key *keys.EdX25519Key) (*Message, error) {
	var message Message
	pk, err := Decrypt(b, &message, key)
	if err != nil {
		return nil, err
	}
	expected := api.NewKey(message.Sender).AsX25519Public()
	if pk.ID() != expected.ID() {
		return nil, errors.Errorf("message sender mismatch")
	}
	return &message, nil
}

// Decrypt value, returning sender public key.
func Decrypt(b []byte, v interface{}, key *keys.EdX25519Key) (*keys.X25519PublicKey, error) {
	box, err := keys.CryptoBoxSealOpen(b, key.X25519Key())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decrypt message")
	}
	if len(box) < 32 {
		return nil, errors.Wrapf(errors.Errorf("not enough bytes"), "failed to decrypt message")
	}
	pk := keys.NewX25519PublicKey(keys.Bytes32(box[:32]))
	encrypted := box[32:]

	decrypted, err := keys.BoxOpen(encrypted, pk, key.X25519Key())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decrypt message")
	}

	if err := msgpack.Unmarshal(decrypted, v); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal message")
	}
	return pk, nil
}

// DecryptMessageFromEvent decrypts a remote Event from Messages.
func DecryptMessageFromEvent(event *events.Event, key *keys.EdX25519Key) (*Message, error) {
	message, err := DecryptMessage(event.Data, key)
	if err != nil {
		return nil, err
	}
	message.RemoteIndex = event.Index
	message.RemoteTimestamp = event.Timestamp
	return message, nil
}
