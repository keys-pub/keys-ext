package client

import (
	"github.com/keys-pub/keys"
	"github.com/vmihailenco/msgpack"
)

func secretBoxMarshal(i interface{}, secretKey *[32]byte) ([]byte, error) {
	b, err := msgpack.Marshal(i)
	if err != nil {
		return nil, err
	}
	return keys.SecretBoxSeal(b, secretKey), nil
}

func secretBoxUnmarshal(b []byte, v interface{}, secretKey *[32]byte) error {
	decrypted, err := keys.SecretBoxOpen(b, secretKey)
	if err != nil {
		return err
	}
	if err := msgpack.Unmarshal(decrypted, v); err != nil {
		return err
	}
	return nil
}
