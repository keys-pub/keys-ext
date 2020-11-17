package api_test

import (
	"encoding/json"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v4"
)

func TestMessageMarshal(t *testing.T) {
	clock := tsutil.NewTestClock()
	msg := &api.Message{
		ID:   "2",
		Prev: "1",
		Text: "hi alice",

		ChannelInfo: &api.ChannelInfo{
			Name:        "test channel",
			Description: "A test channel.",
		},

		Timestamp: clock.NowMillis(),

		// Non-marshalled fields
		Sender:          keys.ID("test"),
		RemoteIndex:     3,
		RemoteTimestamp: clock.NowMillis(),
	}

	b, err := msgpack.Marshal(msg)
	require.NoError(t, err)
	expected := `([]uint8) (len=91 cap=140) {
 00000000  85 a2 69 64 a1 32 a4 70  72 65 76 a1 31 a2 74 73  |..id.2.prev.1.ts|
 00000010  d3 00 00 01 1f 71 fb 04  51 a4 74 65 78 74 a8 68  |.....q..Q.text.h|
 00000020  69 20 61 6c 69 63 65 ab  63 68 61 6e 6e 65 6c 49  |i alice.channelI|
 00000030  6e 66 6f 82 a4 6e 61 6d  65 ac 74 65 73 74 20 63  |nfo..name.test c|
 00000040  68 61 6e 6e 65 6c a4 64  65 73 63 af 41 20 74 65  |hannel.desc.A te|
 00000050  73 74 20 63 68 61 6e 6e  65 6c 2e                 |st channel.|
}
`
	require.Equal(t, expected, spew.Sdump(b))

	b, err = json.MarshalIndent(msg, "", "  ")
	require.NoError(t, err)
	expected = `{
  "id": "2",
  "prev": "1",
  "ts": 1234567890001,
  "text": "hi alice",
  "channelInfo": {
    "name": "test channel",
    "desc": "A test channel."
  }
}`
	require.Equal(t, expected, string(b))
}
