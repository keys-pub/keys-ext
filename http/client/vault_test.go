package client_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/keys-pub/keys"
	httpclient "github.com/keys-pub/keys-ext/http/client"
	"github.com/stretchr/testify/require"
)

func TestVault(t *testing.T) {
	// api.SetLogger(NewLogger(DebugLevel))
	var err error
	env, closeFn := newEnv(t) // client.NewLogger(client.DebugLevel)
	defer closeFn()

	client := newTestClient(t, env)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	exists, err := client.VaultExists(context.TODO(), alice)
	require.NoError(t, err)
	require.False(t, exists)

	event1 := httpclient.NewEvent("/col1/key1", []byte("test1"), nil)
	event2 := httpclient.NewEvent("/col1/key2", []byte("test2"), event1)
	events := []*httpclient.Event{event1, event2}
	err = client.VaultSend(context.TODO(), alice, events)
	require.NoError(t, err)

	vault, err := client.Vault(context.TODO(), alice)
	require.NoError(t, err)
	require.Equal(t, 2, len(vault.Events))
	require.Equal(t, []byte("test1"), vault.Events[0].Data)
	require.Equal(t, []byte("test2"), vault.Events[1].Data)

	exists, err = client.VaultExists(context.TODO(), alice)
	require.NoError(t, err)
	require.True(t, exists)

	event3 := httpclient.NewEvent("/col1/key3", []byte("test3"), event2)
	event4a := httpclient.NewEvent("/col1/key4", []byte("test4.1"), event3)
	event4b := httpclient.NewEvent("/col1/key4", []byte("test4.2"), event4a)
	event5 := httpclient.NewEvent("/col1/key5", []byte("test5"), event4b)
	events2 := []*httpclient.Event{event3, event4a, event4b, event5}

	err = client.VaultSend(context.TODO(), alice, events2)
	require.NoError(t, err)

	vault, err = client.Vault(context.TODO(), alice, httpclient.VaultIndex(vault.Index))
	require.NoError(t, err)
	require.Equal(t, 4, len(vault.Events))
	require.Equal(t, "/col1/key3", vault.Events[0].Path)
	require.Equal(t, []byte("test3"), vault.Events[0].Data)
	require.Equal(t, "/col1/key4", vault.Events[1].Path)
	require.Equal(t, []byte("test4.1"), vault.Events[1].Data)
	require.Equal(t, "/col1/key4", vault.Events[2].Path)
	require.Equal(t, []byte("test4.2"), vault.Events[2].Data)
	require.Equal(t, "/col1/key5", vault.Events[3].Path)
	require.Equal(t, []byte("test5"), vault.Events[3].Data)

	err = client.VaultDelete(context.TODO(), alice)
	require.NoError(t, err)

	exists, err = client.VaultExists(context.TODO(), alice)
	require.NoError(t, err)
	require.False(t, exists)

	err = client.VaultSend(context.TODO(), alice, events)
	require.EqualError(t, err, "vault was deleted (404)")

	err = client.VaultDelete(context.TODO(), alice)
	require.EqualError(t, err, "vault was deleted (404)")
}

func TestVaultMax(t *testing.T) {
	// api.SetLogger(NewLogger(DebugLevel))
	env, closeFn := newEnv(t) // client.NewLogger(client.DebugLevel)
	defer closeFn()

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	testVaultMax(t, env, alice)
}

func testVaultMax(t *testing.T, env *env, key *keys.EdX25519Key) {
	var err error
	client := newTestClient(t, env)

	exists, err := client.VaultExists(context.TODO(), key)
	require.NoError(t, err)
	require.False(t, exists)

	events := make([]*httpclient.Event, 0, 1000)
	for i := 0; i < 1000; i++ {
		event := httpclient.NewEvent(fmt.Sprintf("/col1/key%d", i), []byte(fmt.Sprintf("test%d", i)), nil)
		events = append(events, event)
	}
	err = client.VaultSend(context.TODO(), key, events)
	require.NoError(t, err)

	vault, err := client.Vault(context.TODO(), key)
	require.NoError(t, err)
	require.Equal(t, 1000, len(vault.Events))
	require.Equal(t, []byte("test0"), vault.Events[0].Data)
	require.Equal(t, []byte("test999"), vault.Events[999].Data)
}
