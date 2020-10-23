package vault_test

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v4"
)

func TestSaveKeyDelete(t *testing.T) {
	var err error
	vlt, closeFn := NewTestVault(t, &TestVaultOptions{Unlock: true})
	defer closeFn()

	sk := keys.GenerateEdX25519Key()
	vk := vault.NewKey(sk, time.Now())
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
	key := vault.NewKey(sk, vlt.Now())
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
	key := vault.NewKey(sk, vlt.Now())
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

	vk := vault.NewKey(alice, clock.Now())
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
	expected = `([]uint8) (len=196 cap=388) {
 00000000  86 a2 69 64 d9 3e 6b 65  78 31 33 32 79 77 38 68  |..id.>kex132yw8h|
 00000010  74 35 70 38 63 65 74 6c  32 6a 6d 76 6b 6e 65 77  |t5p8cetl2jmvknew|
 00000020  6a 61 77 74 39 78 77 7a  64 6c 72 6b 32 70 79 78  |jawt9xwzdlrk2pyx|
 00000030  6c 6e 77 6a 79 71 72 64  71 30 64 61 77 71 71 70  |lnwjyqrdq0dawqqp|
 00000040  68 30 37 37 a3 64 61 74  c4 40 01 01 01 01 01 01  |h077.dat.@......|
 00000050  01 01 01 01 01 01 01 01  01 01 01 01 01 01 01 01  |................|
 00000060  01 01 01 01 01 01 01 01  01 01 8a 88 e3 dd 74 09  |..............t.|
 00000070  f1 95 fd 52 db 2d 3c ba  5d 72 ca 67 09 bf 1d 94  |...R.-<.]r.g....|
 00000080  12 1b f3 74 88 01 b4 0f  6f 5c a3 74 79 70 a8 65  |...t....o\.typ.e|
 00000090  64 78 32 35 35 31 39 a5  6e 6f 74 65 73 aa 74 65  |dx25519.notes.te|
 000000a0  73 74 20 6e 6f 74 65 73  a3 63 74 73 d7 ff 00 3d  |st notes.cts...=|
 000000b0  09 00 49 96 02 d2 a3 75  74 73 d7 ff 00 3d 09 00  |..I....uts...=..|
 000000c0  49 96 02 d2                                       |I...|
}
`
	require.Equal(t, expected, spew.Sdump(b))
}

func TestMarshalEdX25519PublicKey(t *testing.T) {
	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	clock := tsutil.NewTestClock()

	vk := vault.NewKey(alice.PublicKey(), clock.Now())
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
	expected = `([]uint8) (len=170 cap=190) {
 00000000  86 a2 69 64 d9 3e 6b 65  78 31 33 32 79 77 38 68  |..id.>kex132yw8h|
 00000010  74 35 70 38 63 65 74 6c  32 6a 6d 76 6b 6e 65 77  |t5p8cetl2jmvknew|
 00000020  6a 61 77 74 39 78 77 7a  64 6c 72 6b 32 70 79 78  |jawt9xwzdlrk2pyx|
 00000030  6c 6e 77 6a 79 71 72 64  71 30 64 61 77 71 71 70  |lnwjyqrdq0dawqqp|
 00000040  68 30 37 37 a3 64 61 74  c4 20 8a 88 e3 dd 74 09  |h077.dat. ....t.|
 00000050  f1 95 fd 52 db 2d 3c ba  5d 72 ca 67 09 bf 1d 94  |...R.-<.]r.g....|
 00000060  12 1b f3 74 88 01 b4 0f  6f 5c a3 74 79 70 ae 65  |...t....o\.typ.e|
 00000070  64 32 35 35 31 39 2d 70  75 62 6c 69 63 a5 6e 6f  |d25519-public.no|
 00000080  74 65 73 aa 74 65 73 74  20 6e 6f 74 65 73 a3 63  |tes.test notes.c|
 00000090  74 73 d7 ff 00 3d 09 00  49 96 02 d2 a3 75 74 73  |ts...=..I....uts|
 000000a0  d7 ff 00 3d 09 00 49 96  02 d2                    |...=..I...|
}
`
	require.Equal(t, expected, spew.Sdump(b))
}

func TestMarshalX25519Key(t *testing.T) {
	alice := keys.NewX25519KeyFromSeed(testSeed(0x01))
	clock := tsutil.NewTestClock()

	vk := vault.NewKey(alice, clock.Now())
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
	expected = `([]uint8) (len=162 cap=190) {
 00000000  86 a2 69 64 d9 3e 6b 62  78 31 35 6e 73 66 39 79  |..id.>kbx15nsf9y|
 00000010  34 6b 32 38 70 38 33 77  74 68 39 33 74 66 37 68  |4k28p83wth93tf7h|
 00000020  61 66 68 76 66 61 6a 70  34 35 64 32 6d 67 65 38  |afhvfajp45d2mge8|
 00000030  30 65 6d 73 34 35 67 7a  30 63 35 67 79 73 35 37  |0ems45gz0c5gys57|
 00000040  63 79 74 6b a3 64 61 74  c4 20 01 01 01 01 01 01  |cytk.dat. ......|
 00000050  01 01 01 01 01 01 01 01  01 01 01 01 01 01 01 01  |................|
 00000060  01 01 01 01 01 01 01 01  01 01 a3 74 79 70 a6 78  |...........typ.x|
 00000070  32 35 35 31 39 a5 6e 6f  74 65 73 aa 74 65 73 74  |25519.notes.test|
 00000080  20 6e 6f 74 65 73 a3 63  74 73 d7 ff 00 3d 09 00  | notes.cts...=..|
 00000090  49 96 02 d2 a3 75 74 73  d7 ff 00 3d 09 00 49 96  |I....uts...=..I.|
 000000a0  02 d2                                             |..|
}
`
	require.Equal(t, expected, spew.Sdump(b))
}

func TestMarshalX25519PublicKey(t *testing.T) {
	alice := keys.NewX25519KeyFromSeed(testSeed(0x01))
	clock := tsutil.NewTestClock()

	vk := vault.NewKey(alice.PublicKey(), clock.Now())
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
	expected = `([]uint8) (len=169 cap=190) {
 00000000  86 a2 69 64 d9 3e 6b 62  78 31 35 6e 73 66 39 79  |..id.>kbx15nsf9y|
 00000010  34 6b 32 38 70 38 33 77  74 68 39 33 74 66 37 68  |4k28p83wth93tf7h|
 00000020  61 66 68 76 66 61 6a 70  34 35 64 32 6d 67 65 38  |afhvfajp45d2mge8|
 00000030  30 65 6d 73 34 35 67 7a  30 63 35 67 79 73 35 37  |0ems45gz0c5gys57|
 00000040  63 79 74 6b a3 64 61 74  c4 20 a4 e0 92 92 b6 51  |cytk.dat. .....Q|
 00000050  c2 78 b9 77 2c 56 9f 5f  a9 bb 13 d9 06 b4 6a b6  |.x.w,V._......j.|
 00000060  8c 9d f9 dc 2b 44 09 f8  a2 09 a3 74 79 70 ad 78  |....+D.....typ.x|
 00000070  32 35 35 31 39 2d 70 75  62 6c 69 63 a5 6e 6f 74  |25519-public.not|
 00000080  65 73 aa 74 65 73 74 20  6e 6f 74 65 73 a3 63 74  |es.test notes.ct|
 00000090  73 d7 ff 00 3d 09 00 49  96 02 d2 a3 75 74 73 d7  |s...=..I....uts.|
 000000a0  ff 00 3d 09 00 49 96 02  d2                       |..=..I...|
}
`
	require.Equal(t, expected, spew.Sdump(b))
}
