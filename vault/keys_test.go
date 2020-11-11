package vault_test

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v4"
)

func TestSaveKeyDelete(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()

	sk := keys.GenerateEdX25519Key()
	vk := api.NewKey(sk)
	require.NoError(t, err)
	out, updated, err := vlt.SaveKey(vk)
	require.NoError(t, err)
	require.False(t, updated)
	require.NotEmpty(t, out.CreatedAt)
	require.NotEmpty(t, out.UpdatedAt)
	key, err := vlt.Key(sk.ID())
	require.NoError(t, err)
	require.NotNil(t, key)
	skOut, err := key.AsEdX25519()
	require.NoError(t, err)
	require.Equal(t, sk.PrivateKey(), skOut.PrivateKey())
	require.Equal(t, sk.PublicKey().Bytes(), skOut.PublicKey().Bytes())

	ok, err := vlt.Delete(sk.ID().String())
	require.NoError(t, err)
	require.True(t, ok)

	out, err = vlt.Key(sk.ID())
	require.NoError(t, err)
	require.Nil(t, out)

	ok, err = vlt.Delete(sk.ID().String())
	require.NoError(t, err)
	require.False(t, ok)
}

func TestStoreConcurrent(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()

	sk := keys.GenerateEdX25519Key()
	key := api.NewKey(sk)
	_, _, err = vlt.SaveKey(key)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < 2000; i++ {
			item, err := vlt.Key(sk.ID())
			require.NoError(t, err)
			require.NotNil(t, item)
		}
		wg.Done()
	}()
	for i := 0; i < 2000; i++ {
		item, err := vlt.Key(sk.ID())
		require.NoError(t, err)
		require.NotNil(t, item)
	}
	wg.Wait()
}

func TestExportImportKey(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()

	sk := keys.GenerateEdX25519Key()
	key := api.NewKey(sk)
	_, _, err = vlt.SaveKey(key)
	require.NoError(t, err)

	password := "testpassword"
	msg, err := vlt.ExportSaltpack(sk.ID(), password)
	require.NoError(t, err)

	vlt2, closeFn2 := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn2()

	out, err := vlt2.ImportSaltpack(msg, "testpassword", false)
	require.NoError(t, err)
	require.Equal(t, sk.ID(), out.ID)
}

func TestMarshalEdX25519Key(t *testing.T) {
	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	clock := tsutil.NewTestClock()

	vk := api.NewKey(alice)
	now := clock.Now()
	vk.CreatedAt = now
	vk.UpdatedAt = now
	vk.Notes = "test notes"

	b, err := json.MarshalIndent(vk, "", "  ")
	require.NoError(t, err)
	expected := `{
  "id": "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
  "data": "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQGKiOPddAnxlf1S2y08ul1yymcJvx2UEhvzdIgBtA9vXA==",
  "type": "edx25519",
  "notes": "test notes",
  "createdAt": "2009-02-13T23:31:30.001Z",
  "updatedAt": "2009-02-13T23:31:30.001Z"
}`
	require.Equal(t, expected, string(b))

	b, err = msgpack.Marshal(vk)
	require.NoError(t, err)
	expected = `([]uint8) (len=210 cap=381) {
 00000000  86 a2 69 64 d9 3e 6b 65  78 31 33 32 79 77 38 68  |..id.>kex132yw8h|
 00000010  74 35 70 38 63 65 74 6c  32 6a 6d 76 6b 6e 65 77  |t5p8cetl2jmvknew|
 00000020  6a 61 77 74 39 78 77 7a  64 6c 72 6b 32 70 79 78  |jawt9xwzdlrk2pyx|
 00000030  6c 6e 77 6a 79 71 72 64  71 30 64 61 77 71 71 70  |lnwjyqrdq0dawqqp|
 00000040  68 30 37 37 a4 64 61 74  61 c4 40 01 01 01 01 01  |h077.data.@.....|
 00000050  01 01 01 01 01 01 01 01  01 01 01 01 01 01 01 01  |................|
 00000060  01 01 01 01 01 01 01 01  01 01 01 8a 88 e3 dd 74  |...............t|
 00000070  09 f1 95 fd 52 db 2d 3c  ba 5d 72 ca 67 09 bf 1d  |....R.-<.]r.g...|
 00000080  94 12 1b f3 74 88 01 b4  0f 6f 5c a4 74 79 70 65  |....t....o\.type|
 00000090  a8 65 64 78 32 35 35 31  39 a5 6e 6f 74 65 73 aa  |.edx25519.notes.|
 000000a0  74 65 73 74 20 6e 6f 74  65 73 a9 63 72 65 61 74  |test notes.creat|
 000000b0  65 64 41 74 d7 ff 00 3d  09 00 49 96 02 d2 a9 75  |edAt...=..I....u|
 000000c0  70 64 61 74 65 64 41 74  d7 ff 00 3d 09 00 49 96  |pdatedAt...=..I.|
 000000d0  02 d2                                             |..|
}
`
	require.Equal(t, expected, spew.Sdump(b))
}

func TestMarshalEdX25519PublicKey(t *testing.T) {
	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	clock := tsutil.NewTestClock()

	vk := api.NewKey(alice.PublicKey())
	now := clock.Now()
	vk.CreatedAt = now
	vk.UpdatedAt = now
	vk.Notes = "test notes"

	b, err := json.MarshalIndent(vk, "", "  ")
	require.NoError(t, err)
	expected := `{
  "id": "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
  "data": "iojj3XQJ8ZX9UtstPLpdcspnCb8dlBIb83SIAbQPb1w=",
  "type": "ed25519-public",
  "notes": "test notes",
  "createdAt": "2009-02-13T23:31:30.001Z",
  "updatedAt": "2009-02-13T23:31:30.001Z"
}`
	require.Equal(t, expected, string(b))

	b, err = msgpack.Marshal(vk)
	require.NoError(t, err)
	expected = `([]uint8) (len=184 cap=190) {
 00000000  86 a2 69 64 d9 3e 6b 65  78 31 33 32 79 77 38 68  |..id.>kex132yw8h|
 00000010  74 35 70 38 63 65 74 6c  32 6a 6d 76 6b 6e 65 77  |t5p8cetl2jmvknew|
 00000020  6a 61 77 74 39 78 77 7a  64 6c 72 6b 32 70 79 78  |jawt9xwzdlrk2pyx|
 00000030  6c 6e 77 6a 79 71 72 64  71 30 64 61 77 71 71 70  |lnwjyqrdq0dawqqp|
 00000040  68 30 37 37 a4 64 61 74  61 c4 20 8a 88 e3 dd 74  |h077.data. ....t|
 00000050  09 f1 95 fd 52 db 2d 3c  ba 5d 72 ca 67 09 bf 1d  |....R.-<.]r.g...|
 00000060  94 12 1b f3 74 88 01 b4  0f 6f 5c a4 74 79 70 65  |....t....o\.type|
 00000070  ae 65 64 32 35 35 31 39  2d 70 75 62 6c 69 63 a5  |.ed25519-public.|
 00000080  6e 6f 74 65 73 aa 74 65  73 74 20 6e 6f 74 65 73  |notes.test notes|
 00000090  a9 63 72 65 61 74 65 64  41 74 d7 ff 00 3d 09 00  |.createdAt...=..|
 000000a0  49 96 02 d2 a9 75 70 64  61 74 65 64 41 74 d7 ff  |I....updatedAt..|
 000000b0  00 3d 09 00 49 96 02 d2                           |.=..I...|
}
`
	require.Equal(t, expected, spew.Sdump(b))
}

func TestMarshalX25519Key(t *testing.T) {
	alice := keys.NewX25519KeyFromSeed(testSeed(0x01))
	clock := tsutil.NewTestClock()

	vk := api.NewKey(alice)
	now := clock.Now()
	vk.CreatedAt = now
	vk.UpdatedAt = now
	vk.Notes = "test notes"

	b, err := json.MarshalIndent(vk, "", "  ")
	require.NoError(t, err)
	expected := `{
  "id": "kbx15nsf9y4k28p83wth93tf7hafhvfajp45d2mge80ems45gz0c5gys57cytk",
  "data": "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE=",
  "type": "x25519",
  "notes": "test notes",
  "createdAt": "2009-02-13T23:31:30.001Z",
  "updatedAt": "2009-02-13T23:31:30.001Z"
}`
	require.Equal(t, expected, string(b))

	b, err = msgpack.Marshal(vk)
	require.NoError(t, err)
	expected = `([]uint8) (len=176 cap=190) {
 00000000  86 a2 69 64 d9 3e 6b 62  78 31 35 6e 73 66 39 79  |..id.>kbx15nsf9y|
 00000010  34 6b 32 38 70 38 33 77  74 68 39 33 74 66 37 68  |4k28p83wth93tf7h|
 00000020  61 66 68 76 66 61 6a 70  34 35 64 32 6d 67 65 38  |afhvfajp45d2mge8|
 00000030  30 65 6d 73 34 35 67 7a  30 63 35 67 79 73 35 37  |0ems45gz0c5gys57|
 00000040  63 79 74 6b a4 64 61 74  61 c4 20 01 01 01 01 01  |cytk.data. .....|
 00000050  01 01 01 01 01 01 01 01  01 01 01 01 01 01 01 01  |................|
 00000060  01 01 01 01 01 01 01 01  01 01 01 a4 74 79 70 65  |............type|
 00000070  a6 78 32 35 35 31 39 a5  6e 6f 74 65 73 aa 74 65  |.x25519.notes.te|
 00000080  73 74 20 6e 6f 74 65 73  a9 63 72 65 61 74 65 64  |st notes.created|
 00000090  41 74 d7 ff 00 3d 09 00  49 96 02 d2 a9 75 70 64  |At...=..I....upd|
 000000a0  61 74 65 64 41 74 d7 ff  00 3d 09 00 49 96 02 d2  |atedAt...=..I...|
}
`
	require.Equal(t, expected, spew.Sdump(b))
}

func TestMarshalX25519PublicKey(t *testing.T) {
	alice := keys.NewX25519KeyFromSeed(testSeed(0x01))
	clock := tsutil.NewTestClock()

	vk := api.NewKey(alice.PublicKey())
	now := clock.Now()
	vk.CreatedAt = now
	vk.UpdatedAt = now
	vk.Notes = "test notes"

	b, err := json.MarshalIndent(vk, "", "  ")
	require.NoError(t, err)
	expected := `{
  "id": "kbx15nsf9y4k28p83wth93tf7hafhvfajp45d2mge80ems45gz0c5gys57cytk",
  "data": "pOCSkrZRwni5dyxWn1+puxPZBrRqtoyd+dwrRAn4ogk=",
  "type": "x25519-public",
  "notes": "test notes",
  "createdAt": "2009-02-13T23:31:30.001Z",
  "updatedAt": "2009-02-13T23:31:30.001Z"
}`
	require.Equal(t, expected, string(b))

	b, err = msgpack.Marshal(vk)
	require.NoError(t, err)
	expected = `([]uint8) (len=183 cap=190) {
 00000000  86 a2 69 64 d9 3e 6b 62  78 31 35 6e 73 66 39 79  |..id.>kbx15nsf9y|
 00000010  34 6b 32 38 70 38 33 77  74 68 39 33 74 66 37 68  |4k28p83wth93tf7h|
 00000020  61 66 68 76 66 61 6a 70  34 35 64 32 6d 67 65 38  |afhvfajp45d2mge8|
 00000030  30 65 6d 73 34 35 67 7a  30 63 35 67 79 73 35 37  |0ems45gz0c5gys57|
 00000040  63 79 74 6b a4 64 61 74  61 c4 20 a4 e0 92 92 b6  |cytk.data. .....|
 00000050  51 c2 78 b9 77 2c 56 9f  5f a9 bb 13 d9 06 b4 6a  |Q.x.w,V._......j|
 00000060  b6 8c 9d f9 dc 2b 44 09  f8 a2 09 a4 74 79 70 65  |.....+D.....type|
 00000070  ad 78 32 35 35 31 39 2d  70 75 62 6c 69 63 a5 6e  |.x25519-public.n|
 00000080  6f 74 65 73 aa 74 65 73  74 20 6e 6f 74 65 73 a9  |otes.test notes.|
 00000090  63 72 65 61 74 65 64 41  74 d7 ff 00 3d 09 00 49  |createdAt...=..I|
 000000a0  96 02 d2 a9 75 70 64 61  74 65 64 41 74 d7 ff 00  |....updatedAt...|
 000000b0  3d 09 00 49 96 02 d2                              |=..I...|
}
`
	require.Equal(t, expected, spew.Sdump(b))
}
