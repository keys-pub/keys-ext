package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v4"
)

// TODO: Test truncated

func TestVault(t *testing.T) {
	// api.SetLogger(NewLogger(DebugLevel))
	var err error
	env, closeFn := newEnv(t) // client.NewLogger(client.DebugLevel)
	defer closeFn()

	cl := newTestClient(t, env)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	exists, err := cl.VaultExists(context.TODO(), alice)
	require.NoError(t, err)
	require.False(t, exists)

	event1 := client.NewVaultEvent("/col1/key1", []byte("test1"))
	event2 := client.NewVaultEvent("/col1/key2", []byte("test2"))
	events := []*client.VaultEvent{event1, event2}
	err = cl.VaultSend(context.TODO(), alice, events)
	require.NoError(t, err)

	vault, err := cl.Vault(context.TODO(), alice)
	require.NoError(t, err)
	require.Equal(t, 2, len(vault.Events))
	require.Equal(t, []byte("test1"), vault.Events[0].Data)
	require.Equal(t, []byte("test2"), vault.Events[1].Data)
	require.False(t, vault.Truncated)

	exists, err = cl.VaultExists(context.TODO(), alice)
	require.NoError(t, err)
	require.True(t, exists)

	event3 := client.NewVaultEvent("/col1/key3", []byte("test3"))
	event4a := client.NewVaultEvent("/col1/key4", []byte("test4.1"))
	event4b := client.NewVaultEvent("/col1/key4", []byte("test4.2"))
	event5 := client.NewVaultEvent("/col1/key5", []byte("test5"))
	events2 := []*client.VaultEvent{event3, event4a, event4b, event5}

	err = cl.VaultSend(context.TODO(), alice, events2)
	require.NoError(t, err)

	vault, err = cl.Vault(context.TODO(), alice, client.VaultIndex(vault.Index))
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

	err = cl.VaultDelete(context.TODO(), alice)
	require.NoError(t, err)

	exists, err = cl.VaultExists(context.TODO(), alice)
	require.NoError(t, err)
	require.False(t, exists)

	err = cl.VaultSend(context.TODO(), alice, events)
	require.EqualError(t, err, "vault was deleted (404)")

	err = cl.VaultDelete(context.TODO(), alice)
	require.EqualError(t, err, "vault was deleted (404)")
}

func TestVaultMax(t *testing.T) {
	// api.SetLogger(NewLogger(DebugLevel))
	env, closeFn := newEnv(t) // client.NewLogger(client.DebugLevel)
	defer closeFn()

	key := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	testVaultMax(t, env, key)
}

func TestVaultMaxFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	env, closeFn := newEnvWithOptions(t, &envOptions{fi: testFirestore(t), clock: tsutil.NewTestClock()})
	defer closeFn()

	key := keys.GenerateEdX25519Key()

	testVaultMax(t, env, key)
}

func testVaultMax(t *testing.T, env *env, key *keys.EdX25519Key) {
	var err error
	cl := newTestClient(t, env)

	exists, err := cl.VaultExists(context.TODO(), key)
	require.NoError(t, err)
	require.False(t, exists)

	events := make([]*client.VaultEvent, 0, 1000)
	for i := 0; i < 1000; i++ {
		event := client.NewVaultEvent(fmt.Sprintf("/col1/key%d", i), []byte(fmt.Sprintf("test%d", i)))
		events = append(events, event)
	}
	err = cl.VaultSend(context.TODO(), key, events)
	require.NoError(t, err)

	vault, err := cl.Vault(context.TODO(), key)
	require.NoError(t, err)
	require.Equal(t, 1000, len(vault.Events))
	require.Equal(t, []byte("test0"), vault.Events[0].Data)
	require.Equal(t, []byte("test999"), vault.Events[999].Data)
}

func TestVaultEventMarshal(t *testing.T) {
	clock := tsutil.NewTestClock()
	event := &client.VaultEvent{
		Path:            "/vault/1",
		Data:            []byte("test"),
		RemoteIndex:     3,
		RemoteTimestamp: clock.Now(),
	}

	b, err := msgpack.Marshal(event)
	require.NoError(t, err)
	expected := `([]uint8) (len=22 cap=64) {
 00000000  82 a1 70 a8 2f 76 61 75  6c 74 2f 31 a3 64 61 74  |..p./vault/1.dat|
 00000010  c4 04 74 65 73 74                                 |..test|
}
`
	require.Equal(t, expected, spew.Sdump(b))

	b, err = json.MarshalIndent(event, "", "  ")
	require.NoError(t, err)
	expected = `{
  "path": "/vault/1",
  "data": "dGVzdA=="
}`
	require.Equal(t, expected, string(b))
}
