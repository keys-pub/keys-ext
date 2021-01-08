package api_test

import (
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/stretchr/testify/require"
)

func TestEncrypt(t *testing.T) {
	type test struct {
		Value string `msgpack:"val"`
	}

	v := &test{Value: "testing"}
	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	bob := keys.NewEdX25519KeyFromSeed(testSeed(0x02))

	b, err := api.Encrypt(v, alice, bob.ID())
	require.NoError(t, err)

	var out test
	pk, err := api.Decrypt(b, &out, bob)
	require.NoError(t, err)
	require.Equal(t, pk, alice.X25519Key().PublicKey())
	require.Equal(t, &out, v)
}

func TestMessageEncrypt(t *testing.T) {
	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	bob := keys.NewEdX25519KeyFromSeed(testSeed(0x02))

	msg := api.NewMessage(alice.ID()).WithText("test message")

	encrypted, err := msg.Encrypt(alice, bob.ID())
	require.NoError(t, err)

	out, err := api.DecryptMessage(encrypted, bob)
	require.NoError(t, err)
	require.Equal(t, msg, out)
	require.Equal(t, alice.ID(), out.Sender)
}
