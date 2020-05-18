package git

import (
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/keyring"
	"github.com/keys-pub/keys/saltpack"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v4"
)

func encryptItem(item *keyring.Item, key *keys.EdX25519Key) ([]byte, error) {
	b, err := msgpack.Marshal(item)
	if err != nil {
		return nil, err
	}
	encrypted, err := saltpack.Signcrypt(b, key, key.ID())
	if err != nil {
		return nil, err
	}
	return encrypted, nil
}

func decryptItem(b []byte, key *keys.EdX25519Key, ks saltpack.KeyStore) (*keyring.Item, error) {
	decrypted, sender, err := saltpack.SigncryptOpen(b, ks)
	if err != nil {
		return nil, err
	}
	if sender == nil {
		return nil, errors.Errorf("no sender")
	}
	// TODO: CHeck sender?
	// if sender.ID() != key.ID() {
	// 	return nil, errors.Errorf("unknown sender %s", sender.ID())
	// }
	var item keyring.Item
	if err := msgpack.Unmarshal(decrypted, &item); err != nil {
		return nil, err
	}
	return &item, nil
}
