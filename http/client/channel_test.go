package client_test

import (
	"context"
	"os"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys-ext/http/client"
	kapi "github.com/keys-pub/keys/api"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestChannel(t *testing.T) {
	env, closeFn := newEnv(t)
	defer closeFn()
	testChannel(t, env, testKeysSeeded())
}

func testChannel(t *testing.T, env *env, tk testKeys) {
	alice, bob, channel, channel2 := tk.alice, tk.bob, tk.channel, tk.channel2

	aliceClient := newTestClient(t, env)
	bobClient := newTestClient(t, env)

	ctx := context.TODO()
	info := &api.ChannelInfo{Name: "test"}
	create1, err := aliceClient.ChannelCreate(ctx, channel, alice, info)
	require.NoError(t, err)
	token1 := &api.ChannelToken{ID: channel.ID(), Token: create1.Channel.Token}

	info2 := &api.ChannelInfo{Name: "test2"}
	create2, err := aliceClient.ChannelCreate(ctx, channel2, alice, info2)
	require.NoError(t, err)
	token2 := &api.ChannelToken{ID: channel2.ID(), Token: create2.Channel.Token}

	channels, err := aliceClient.ChannelsStatus(ctx, token1, token2)
	require.NoError(t, err)
	expected := []*api.ChannelStatus{
		&api.ChannelStatus{
			ID:        keys.ID("kex1tan3x22v8nc6s98gmr9q3zwmy0ngywm4yja0zdylh37e752jj3dsur2s3g"),
			Index:     int64(1),
			Timestamp: int64(1234567890011),
		},
		&api.ChannelStatus{ID: keys.ID("kex1fzlrdfy4wlyaturcqkfq92ywj7lft9awtdg70d2yftzhspmc45qsvghhep"),
			Index:     int64(1),
			Timestamp: int64(1234567890004),
		},
	}
	require.Equal(t, expected, channels)

	// Invites
	invite := &api.ChannelInvite{
		Channel:   channel.ID(),
		Info:      info,
		Recipient: bob.ID(),
		Sender:    alice.ID(),
		Key:       kapi.NewKey(channel),
		Token:     create1.Channel.Token,
	}
	_, err = aliceClient.InviteToChannel(ctx, invite, alice, "")
	require.EqualError(t, err, "auth failed (403)")

	bobToken := client.GenerateToken()
	err = bobClient.DropAuth(ctx, bob, bobToken)
	require.NoError(t, err)

	_, err = aliceClient.InviteToChannel(ctx, invite, alice, bobToken)
	require.NoError(t, err)

	invites, err := bobClient.ChannelInvites(ctx, bob, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(invites.Invites))
	require.Equal(t, invite, invites.Invites[0])
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
