package vault_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v4"
)

func TestProvisions(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()

	provisions, err := vlt.Provisions()
	require.NoError(t, err)
	require.Equal(t, 1, len(provisions))
	require.Equal(t, "ySymDh5DDuJo21ydVJdyuxcDTgYUJMin4PZQzSUBums", provisions[0].ID)

	// TODO: Test provisioning the same key twice?
}

func TestProvisionSave(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, nil)
	defer closeFn()

	id := encoding.MustEncode(bytes.Repeat([]byte{0x01}, 32), encoding.Base62)
	provision := &vault.Provision{
		ID: id,
	}
	err = vlt.ProvisionSave(provision)
	require.NoError(t, err)

	provisions, err := vlt.Provisions()
	require.NoError(t, err)
	require.Equal(t, 1, len(provisions))
	require.Equal(t, "0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29", provisions[0].ID)
}

func TestProvisionMarshal(t *testing.T) {
	var err error

	clock := tsutil.NewTestClock()
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Clock: clock, Unlock: true})
	defer closeFn()

	id := encoding.MustEncode(bytes.Repeat([]byte{0x01}, 32), encoding.Base62)
	salt := bytes.Repeat([]byte{0x02}, 32)
	provision := &vault.Provision{
		ID:        id,
		Type:      vault.PasswordAuth,
		AAGUID:    "123",
		Salt:      salt,
		NoPin:     true,
		CreatedAt: clock.Now(),
	}
	key := keys.Rand32()
	err = vlt.Provision(key, provision)
	require.NoError(t, err)

	provisions, err := vlt.Provisions()
	require.NoError(t, err)
	require.Equal(t, 2, len(provisions))
	out := provisions[1]
	require.Equal(t, provision.ID, out.ID)
	require.Equal(t, provision.Salt, out.Salt)
	require.Equal(t, provision.AAGUID, out.AAGUID)
	require.Equal(t, provision.NoPin, out.NoPin)
	require.Equal(t, provision.Type, out.Type)
	require.Equal(t, provision.CreatedAt.UTC(), out.CreatedAt.UTC())

	b, err := msgpack.Marshal(provision)
	require.NoError(t, err)
	expected := `([]uint8) (len=134 cap=267) {
 00000000  86 a2 69 64 d9 2b 30 45  6c 36 58 46 58 77 73 55  |..id.+0El6XFXwsU|
 00000010  46 44 38 4a 32 76 47 78  73 61 62 6f 57 37 72 5a  |FD8J2vGxsaboW7rZ|
 00000020  59 6e 51 52 42 50 35 64  39 65 72 77 52 77 64 32  |YnQRBP5d9erwRwd2|
 00000030  39 a4 74 79 70 65 a8 70  61 73 73 77 6f 72 64 a3  |9.type.password.|
 00000040  63 74 73 d7 ff 00 7a 12  00 49 96 02 d2 a6 61 61  |cts...z..I....aa|
 00000050  67 75 69 64 a3 31 32 33  a4 73 61 6c 74 c4 20 02  |guid.123.salt. .|
 00000060  02 02 02 02 02 02 02 02  02 02 02 02 02 02 02 02  |................|
 00000070  02 02 02 02 02 02 02 02  02 02 02 02 02 02 02 a5  |................|
 00000080  6e 6f 70 69 6e c3                                 |nopin.|
}
`
	require.Equal(t, expected, spew.Sdump(b))

	b, err = json.MarshalIndent(provision, "", "  ")
	require.NoError(t, err)
	expectedJSON := `{
  "id": "0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29",
  "type": "password",
  "cts": "2009-02-13T23:31:30.002Z",
  "aaguid": "123",
  "salt": "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI=",
  "nopin": true
}`
	require.Equal(t, expectedJSON, string(b))
}
