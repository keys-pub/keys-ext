package client_test

import (
	"context"
	"os"
	"testing"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestChannel(t *testing.T) {
	env, closeFn := newEnv(t)
	defer closeFn()
	testChannel(t, env, testKeysSeeded())
}

func TestChannelFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	env, closeFn := newEnvWithOptions(t, &envOptions{fi: testFirestore(t), clock: tsutil.NewTestClock()})
	defer closeFn()

	env.logger = client.NewLogger(client.DebugLevel)

	testChannel(t, env, testKeysRandom())
}

func testChannel(t *testing.T, env *env, tk testKeys) {
	alice, channel := tk.alice, tk.channel

	aliceClient := newTestClient(t, env)
	// bobClient := newTestClient(t, env)

	ctx := context.TODO()
	info := &api.ChannelInfo{Name: "test"}

	_, err := aliceClient.ChannelCreate(ctx, channel, alice, info)
	require.NoError(t, err)

	// _, err = aliceClient.InviteToChannel(ctx, channel, info, alice, bob.ID())
	// require.NoError(t, err)
}
