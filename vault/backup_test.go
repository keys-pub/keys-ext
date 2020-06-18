package vault_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestBackupRestore(t *testing.T) {
	var err error
	clock := tsutil.NewClock()

	st := vault.NewMem()

	vlt := vault.New(st)
	err = vlt.UnlockWithPassword("testpassword", true)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		err := st.Set(fmt.Sprintf("item%d", i), []byte(fmt.Sprintf("value%d", i)))
		require.NoError(t, err)
	}

	tmpFile := keys.RandTempPath() + ".tgz"
	defer func() { _ = os.Remove(tmpFile) }()

	err = vault.Backup(tmpFile, st, clock.Now())
	require.NoError(t, err)

	st2 := vault.NewMem()
	err = vault.Restore(tmpFile, st2)
	require.NoError(t, err)
	testEqualKeyrings(t, st, st2)

	vlt2 := vault.New(st2)
	err = vlt2.UnlockWithPassword("testpassword", false)
	require.NoError(t, err)
}

func testEqualKeyrings(t *testing.T, st1 vault.Store, st2 vault.Store) {
	iter1, err := st1.Documents()
	require.NoError(t, err)
	docs1, err := ds.DocumentsFromIterator(iter1)
	require.NoError(t, err)
	iter2, err := st2.Documents()
	require.NoError(t, err)
	docs2, err := ds.DocumentsFromIterator(iter2)
	require.NoError(t, err)

	require.Equal(t, len(docs1), len(docs2))

	for i := 0; i < len(docs1); i++ {
		b1, err := st1.Get(docs1[i].Path)
		require.NoError(t, err)
		b2, err := st2.Get(docs2[i].Path)
		require.NoError(t, err)
		require.Equal(t, b1, b2)
	}
}
