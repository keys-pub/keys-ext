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
		Content: &api.Content{
			Data: []byte("hi alice"),
			Type: api.UTF8Content,
		},
		CreatedAt: clock.Now(),

		Sender:          keys.ID("test"),
		RemoteIndex:     3,
		RemoteTimestamp: clock.Now(),
	}

	b, err := msgpack.Marshal(msg)
	require.NoError(t, err)
	expected := `([]uint8) (len=67 cap=136) {
 00000000  84 a2 69 64 a1 32 a4 70  72 65 76 a1 31 a7 63 6f  |..id.2.prev.1.co|
 00000010  6e 74 65 6e 74 82 a4 64  61 74 61 c4 08 68 69 20  |ntent..data..hi |
 00000020  61 6c 69 63 65 a4 74 79  70 65 a4 75 74 66 38 a9  |alice.type.utf8.|
 00000030  63 72 65 61 74 65 64 41  74 d7 ff 00 3d 09 00 49  |createdAt...=..I|
 00000040  96 02 d2                                          |...|
}
`
	require.Equal(t, expected, spew.Sdump(b))

	b, err = json.MarshalIndent(msg, "", "  ")
	require.NoError(t, err)
	expected = `{
  "id": "2",
  "prev": "1",
  "content": {
    "data": "aGkgYWxpY2U=",
    "type": "utf8"
  },
  "createdAt": "2009-02-13T23:31:30.001Z"
}`
	require.Equal(t, expected, string(b))
}
