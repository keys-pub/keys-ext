package vault

import (
	"github.com/keys-pub/keys"
	"golang.org/x/crypto/nacl/secretbox"
)

func secretBoxSeal(b []byte, secretKey *[32]byte) []byte {
	nonce := keys.Rand24()
	return secretBoxSealWithNonce(b, nonce, secretKey)
}

func secretBoxSealWithNonce(b []byte, nonce *[24]byte, secretKey *[32]byte) []byte {
	encrypted := secretbox.Seal(nil, b, nonce, secretKey)
	encrypted = append(nonce[:], encrypted...)
	return encrypted
}

func secretBoxOpen(encrypted []byte, secretKey *[32]byte) ([]byte, bool) {
	if secretKey == nil {
		return nil, false
	}
	if len(encrypted) < 24 {
		return nil, false
	}
	var nonce [24]byte
	copy(nonce[:], encrypted[:24])
	encrypted = encrypted[24:]

	return secretbox.Open(nil, encrypted, &nonce, secretKey)
}
