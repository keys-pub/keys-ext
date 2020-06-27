package vault_test

import (
	"sync"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestX25519KeyItem(t *testing.T) {
	key := keys.GenerateX25519Key()
	out, err := vault.KeyForItem(vault.ItemForKey(key))
	require.NoError(t, err)
	require.Equal(t, key.ID(), out.ID())
}

func TestX25519PublicKeyItem(t *testing.T) {
	key := keys.GenerateX25519Key()
	out, err := vault.KeyForItem(vault.ItemForKey(key.ID()))
	require.NoError(t, err)
	require.Equal(t, key.ID(), out.ID())
}

func TestEdX25519KeyItem(t *testing.T) {
	key := keys.GenerateEdX25519Key()
	out, err := vault.KeyForItem(vault.ItemForKey(key))
	require.NoError(t, err)
	require.Equal(t, key.ID(), out.ID())
}

func TestEdX25519PublicKeyItem(t *testing.T) {
	key := keys.GenerateEdX25519Key()
	out, err := vault.KeyForItem(vault.ItemForKey(key.ID()))
	require.NoError(t, err)
	require.Equal(t, key.ID(), out.ID())
}

func TestSaveKeyDelete(t *testing.T) {
	var err error
	db, closeFn := newTestVaultDB(t)
	defer closeFn()
	vlt := vault.New(db)
	err = vlt.Setup(keys.Rand32(), vault.NewProvision(vault.UnknownAuth))
	require.NoError(t, err)

	sk := keys.GenerateEdX25519Key()
	err = vlt.SaveKey(sk)
	require.NoError(t, err)
	out, err := vlt.EdX25519Key(sk.ID())
	require.NoError(t, err)
	require.Equal(t, sk.PrivateKey(), out.PrivateKey())
	require.Equal(t, sk.PublicKey().Bytes(), out.PublicKey().Bytes())

	ok, err := vlt.Delete(sk.ID().String())
	require.NoError(t, err)
	require.True(t, ok)

	out, err = vlt.EdX25519Key(sk.ID())
	require.NoError(t, err)
	require.Nil(t, out)

	ok, err = vlt.Delete(sk.ID().String())
	require.NoError(t, err)
	require.False(t, ok)
}

func TestEdX25519Key(t *testing.T) {
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))
	var err error
	clock := tsutil.NewClock()
	vlt := newTestVaultUnlocked(t, clock)
	sk := keys.GenerateEdX25519Key()

	err = vlt.SaveKey(sk)
	require.NoError(t, err)
	skOut, err := vlt.EdX25519Key(sk.ID())
	require.NoError(t, err)
	require.Equal(t, sk.PrivateKey()[:], skOut.PrivateKey()[:])
	require.Equal(t, sk.PublicKey().Bytes()[:], skOut.PublicKey().Bytes()[:])

	sks, err := vlt.EdX25519Keys()
	require.NoError(t, err)
	require.Equal(t, 1, len(sks))
	require.Equal(t, sk.Seed()[:], sks[0].Seed()[:])

	spk := keys.GenerateEdX25519Key().PublicKey()
	err = vlt.SaveKey(spk)
	require.NoError(t, err)
	skOut, err = vlt.EdX25519Key(spk.ID())
	require.NoError(t, err)
	require.Nil(t, skOut)
}

func TestEdX25519PublicKey(t *testing.T) {
	var err error
	clock := tsutil.NewClock()
	vlt := newTestVaultUnlocked(t, clock)

	sk := keys.GenerateEdX25519Key()
	err = vlt.SaveKey(sk)
	require.NoError(t, err)

	spkConv, err := vlt.EdX25519PublicKey(sk.PublicKey().X25519PublicKey().ID())
	require.NoError(t, err)
	require.Equal(t, sk.PublicKey().Bytes(), spkConv.Bytes())

	spk := keys.GenerateEdX25519Key().PublicKey()
	err = vlt.SaveKey(spk)
	require.NoError(t, err)

	spkConv2, err := vlt.EdX25519PublicKey(spk.X25519PublicKey().ID())
	require.NoError(t, err)
	require.Equal(t, spk.Bytes(), spkConv2.Bytes())
}

func TestX25519Key(t *testing.T) {
	var err error
	clock := tsutil.NewClock()
	vlt := newTestVaultUnlocked(t, clock)

	bk := keys.GenerateX25519Key()
	err = vlt.SaveKey(bk)
	require.NoError(t, err)
	bkOut, err := vlt.X25519Key(bk.ID())
	require.NoError(t, err)
	require.Equal(t, bk.PrivateKey()[:], bkOut.PrivateKey()[:])
	require.Equal(t, bk.PublicKey().Bytes()[:], bkOut.PublicKey().Bytes()[:])

	bpk := keys.GenerateX25519Key().PublicKey()
	err = vlt.SaveKey(bpk)
	require.NoError(t, err)
	bkOut, err = vlt.X25519Key(bpk.ID())
	require.NoError(t, err)
	require.Nil(t, bkOut)
}

func TestKeys(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	var err error
	clock := tsutil.NewClock()
	vlt := newTestVaultUnlocked(t, clock)

	sk := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	err = vlt.SaveKey(sk)
	require.NoError(t, err)

	sk2 := keys.NewEdX25519KeyFromSeed(testSeed(0x02))
	err = vlt.SaveKey(sk2.PublicKey())
	require.NoError(t, err)

	bk := keys.NewX25519KeyFromSeed(testSeed(0x01))
	err = vlt.SaveKey(bk)
	require.NoError(t, err)

	bk2 := keys.NewX25519KeyFromSeed(testSeed(0x02))
	err = vlt.SaveKey(bk2.PublicKey())
	require.NoError(t, err)

	out, err := vlt.Keys()
	require.NoError(t, err)
	require.Equal(t, 4, len(out))

	out, err = vlt.Keys(vault.Keys.Types(keys.X25519, keys.X25519Public))
	require.NoError(t, err)
	require.Equal(t, 2, len(out))
	require.Equal(t, bk.ID(), out[0].ID())
	require.Equal(t, bk2.ID(), out[1].ID())

	out, err = vlt.Keys(vault.Keys.Types(keys.X25519))
	require.NoError(t, err)
	require.Equal(t, 1, len(out))
	require.Equal(t, bk.ID(), out[0].ID())
}

func TestStoreConcurrent(t *testing.T) {
	var err error
	clock := tsutil.NewClock()
	vlt := newTestVaultUnlocked(t, clock)

	sk := keys.GenerateEdX25519Key()
	err = vlt.SaveKey(sk)
	require.NoError(t, err)

	skOut, err := vlt.EdX25519Key(sk.ID())
	require.NoError(t, err)
	require.Equal(t, sk.Seed()[:], skOut.Seed()[:])

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < 2000; i++ {
			skOut, err := vlt.EdX25519Key(sk.ID())
			require.NoError(t, err)
			require.Equal(t, sk.Seed()[:], skOut.Seed()[:])
		}
		wg.Done()
	}()
	for i := 0; i < 2000; i++ {
		skOut, err := vlt.EdX25519Key(sk.ID())
		require.NoError(t, err)
		require.Equal(t, sk.Seed()[:], skOut.Seed()[:])
	}
	wg.Wait()
}

func TestExportImportKey(t *testing.T) {
	var err error
	clock := tsutil.NewClock()
	vlt := newTestVaultUnlocked(t, clock)

	sk := keys.GenerateEdX25519Key()
	err = vlt.SaveKey(sk)
	require.NoError(t, err)

	password := "testpassword"
	msg, err := vlt.ExportSaltpack(sk.ID(), password)
	require.NoError(t, err)

	vlt2 := newTestVaultUnlocked(t, clock)

	key, err := vlt2.ImportSaltpack(msg, "testpassword", false)
	require.NoError(t, err)
	require.Equal(t, sk.ID(), key.ID())
}

func TestUnknownKey(t *testing.T) {
	var err error
	clock := tsutil.NewClock()
	vlt := newTestVaultUnlocked(t, clock)

	key, err := vlt.Key(keys.RandID("kex"))
	require.NoError(t, err)
	require.Nil(t, key)
}
