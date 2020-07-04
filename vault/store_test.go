package vault_test

import (
	"testing"

	"github.com/keys-pub/keys/docs"

	"github.com/keys-pub/keys-ext/vault"
	"github.com/stretchr/testify/require"
)

func TestStoreMem(t *testing.T) {
	testStore(t, vault.NewMem())
}

func TestStoreDB(t *testing.T) {
	db, closeFn := newTestVaultDB(t)
	defer closeFn()
	testStore(t, db)
}

func testStore(t *testing.T, st vault.Store) {
	var err error

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

	out, err := st.Documents(docs.Prefix("/col1"))
	require.NoError(t, err)
	require.Equal(t, 2, len(out))
	require.Equal(t, "/col1/key1", out[0].Path)
	require.Equal(t, []byte("val1"), out[0].Data)
	require.Equal(t, "/col1/key2", out[1].Path)
	require.Equal(t, []byte("val2"), out[1].Data)

	out, err = st.Documents(docs.Prefix("/col1"), docs.Limit(1))
	require.NoError(t, err)
	require.Equal(t, 1, len(out))
	require.Equal(t, "/col1/key1", out[0].Path)
	require.Equal(t, []byte("val1"), out[0].Data)
}
