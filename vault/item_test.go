package vault_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v4"
)

func TestItem(t *testing.T) {
	clock := tsutil.NewTestClock()
	item := vault.NewItem("account1", []byte("password"), "passphrase", clock.Now())

	b, err := msgpack.Marshal(item)
	require.NoError(t, err)
	expected := `([]uint8) (len=56 cap=64) {
 00000000  84 a2 69 64 a8 61 63 63  6f 75 6e 74 31 a3 64 61  |..id.account1.da|
 00000010  74 c4 08 70 61 73 73 77  6f 72 64 a3 74 79 70 aa  |t..password.typ.|
 00000020  70 61 73 73 70 68 72 61  73 65 a3 63 74 73 d7 ff  |passphrase.cts..|
 00000030  00 3d 09 00 49 96 02 d2                           |.=..I...|
}
`
	require.Equal(t, expected, spew.Sdump(b))
}

func TestLargeItems(t *testing.T) {
	var err error
	const maxID = 254
	const maxType = 32
	const maxData = 2048

	vlt := vault.New(vault.NewMem())

	key := keys.Bytes32(bytes.Repeat([]byte{0x01}, 32))
	provision := vault.NewProvision(vault.UnknownAuth)
	err = vlt.Setup(key, provision)
	require.NoError(t, err)

	id := string(bytes.Repeat([]byte("a"), maxID))
	largeID := string(bytes.Repeat([]byte("a"), maxID+1))
	typ := string(bytes.Repeat([]byte("t"), maxType))
	largeType := string(bytes.Repeat([]byte("a"), maxType+1))

	large := keys.RandBytes(maxData + 1)
	err = vlt.Set(vault.NewItem(id, large, typ, time.Now()))
	require.EqualError(t, err, "item value is too large")

	err = vlt.Set(vault.NewItem(largeID, []byte{0x01}, typ, time.Now()))
	require.EqualError(t, err, "item value is too large")
	err = vlt.Set(vault.NewItem(id, []byte{0x01}, largeType, time.Now()))
	require.EqualError(t, err, "item value is too large")

	b := bytes.Repeat([]byte{0x01}, maxData)
	err = vlt.Set(vault.NewItem(id, b, typ, time.Now()))
	require.NoError(t, err)

	item, err := vlt.Get(id)
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, b, item.Data)
}
