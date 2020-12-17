package api_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v4"
)

func testSeed(b byte) *[32]byte {
	return keys.Bytes32(bytes.Repeat([]byte{b}, 32))
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

func TestMessageMarshal(t *testing.T) {
	clock := tsutil.NewTestClock()

	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))

	msg := &api.Message{
		ID:     "2",
		Prev:   "1",
		Text:   "hi alice",
		Sender: alice.ID(),

		ChannelInfo: &api.ChannelInfo{
			Name:        "test channel",
			Description: "A test channel.",
		},

		Timestamp: clock.NowMillis(),

		// Non-marshalled fields
		RemoteIndex:     3,
		RemoteTimestamp: clock.NowMillis(),
	}

	b, err := msgpack.Marshal(msg)
	require.NoError(t, err)
	expected := `([]uint8) (len=162 cap=190) {
 00000000  86 a2 69 64 a1 32 a4 70  72 65 76 a1 31 a2 74 73  |..id.2.prev.1.ts|
 00000010  d3 00 00 01 1f 71 fb 04  51 a6 73 65 6e 64 65 72  |.....q..Q.sender|
 00000020  d9 3e 6b 65 78 31 33 32  79 77 38 68 74 35 70 38  |.>kex132yw8ht5p8|
 00000030  63 65 74 6c 32 6a 6d 76  6b 6e 65 77 6a 61 77 74  |cetl2jmvknewjawt|
 00000040  39 78 77 7a 64 6c 72 6b  32 70 79 78 6c 6e 77 6a  |9xwzdlrk2pyxlnwj|
 00000050  79 71 72 64 71 30 64 61  77 71 71 70 68 30 37 37  |yqrdq0dawqqph077|
 00000060  a4 74 65 78 74 a8 68 69  20 61 6c 69 63 65 ab 63  |.text.hi alice.c|
 00000070  68 61 6e 6e 65 6c 49 6e  66 6f 82 a4 6e 61 6d 65  |hannelInfo..name|
 00000080  ac 74 65 73 74 20 63 68  61 6e 6e 65 6c a4 64 65  |.test channel.de|
 00000090  73 63 af 41 20 74 65 73  74 20 63 68 61 6e 6e 65  |sc.A test channe|
 000000a0  6c 2e                                             |l.|
}
`
	require.Equal(t, expected, spew.Sdump(b))

	b, err = json.MarshalIndent(msg, "", "  ")
	require.NoError(t, err)
	expected = `{
  "id": "2",
  "prev": "1",
  "ts": 1234567890001,
  "sender": "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
  "text": "hi alice",
  "channelInfo": {
    "name": "test channel",
    "desc": "A test channel."
  }
}`
	require.Equal(t, expected, string(b))
}
