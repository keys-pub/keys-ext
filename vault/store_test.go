package vault_test

import (
	"os"
	"testing"

	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys-ext/vault/leveldb"
	"github.com/stretchr/testify/require"
)

func TestStoreMem(t *testing.T) {
	mem := vault.NewMem()
	testStore(t, mem)
}

func TestStoreLevelDB(t *testing.T) {
	path := testPath()
	db := leveldb.New(path)
	defer func() {
		err := db.Close()
		require.NoError(t, err)
		_ = os.RemoveAll(path)
	}()
	testStore(t, db)
}

func testStore(t *testing.T, st vault.Store) {
	var err error

	_, err = st.Get("/col1/key1")
	require.EqualError(t, err, "vault not open")

	err = st.Open()
	require.NoError(t, err)
	defer st.Close()

	err = st.Open()
	require.EqualError(t, err, "vault already open")

	b, err := st.Get("/col1/key1")
	require.NoError(t, err)
	require.Nil(t, b)

	err = st.Set("/col1/key1", []byte("val1"))
	require.NoError(t, err)
	err = st.Set("/col1/key2", []byte("val2"))
	require.NoError(t, err)

	b, err = st.Get("/col1/key1")
	require.NoError(t, err)
	require.Equal(t, []byte("val1"), b)

	out, err := st.List(&vault.ListOptions{Prefix: "/col1"})
	require.NoError(t, err)
	require.Equal(t, 2, len(out))
	require.Equal(t, "/col1/key1", out[0].Path)
	require.Equal(t, []byte("val1"), out[0].Data)
	require.Equal(t, "/col1/key2", out[1].Path)
	require.Equal(t, []byte("val2"), out[1].Data)

	out, err = st.List(&vault.ListOptions{Prefix: "/col1", Limit: 1})
	require.NoError(t, err)
	require.Equal(t, 1, len(out))
	require.Equal(t, "/col1/key1", out[0].Path)
	require.Equal(t, []byte("val1"), out[0].Data)

	err = st.Reset()
	require.NoError(t, err)

	b, err = st.Get("/col1/key1")
	require.NoError(t, err)
	require.Nil(t, b)
}
