package client_test

import (
	"context"
	"os"
	"testing"

	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/saltpack"
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
	alice, bob, channel := tk.alice, tk.bob, tk.channel

	aliceClient := newTestClient(t, env)
	// bobClient := newTestClient(t, env)

	ctx := context.TODO()
	info := &api.ChannelInfo{Name: "test"}

	_, err := aliceClient.ChannelCreate(ctx, channel, alice, info)
	require.NoError(t, err)

	_, err = aliceClient.InviteToChannel(ctx, channel, info, alice, bob.ID())
	require.NoError(t, err)

	invites, err := aliceClient.ChannelInvites(ctx, channel, alice)
	require.NoError(t, err)
	require.Equal(t, 1, len(invites))
	require.Equal(t, bob.ID(), invites[0].Recipient)
	require.Equal(t, channel.ID(), invites[0].Channel)

	ck, pk, err := invites[0].DecryptKey(saltpack.NewKeyring(bob))
	require.NoError(t, err)
	require.Equal(t, alice.ID(), pk)
	require.Equal(t, channel, ck)

	var outInfo *api.ChannelInfo
	pk, err = api.Decrypt(invites[0].Info, &outInfo, saltpack.NewKeyring(bob))
	require.NoError(t, err)
	require.Equal(t, alice.ID(), pk)
	require.Equal(t, outInfo, info)

	invites, err = aliceClient.UserChannelInvites(ctx, bob)
	require.NoError(t, err)
	require.Equal(t, 1, len(invites))
	require.Equal(t, bob.ID(), invites[0].Recipient)
	require.Equal(t, channel.ID(), invites[0].Channel)

	_, err = aliceClient.ChannelUninvite(ctx, channel, alice, bob.ID())
	require.NoError(t, err)

	invites, err = aliceClient.ChannelInvites(ctx, channel, alice)
	require.NoError(t, err)
	require.Equal(t, 0, len(invites))

	invites, err = aliceClient.UserChannelInvites(ctx, bob)
	require.NoError(t, err)
	require.Equal(t, 0, len(invites))
}
