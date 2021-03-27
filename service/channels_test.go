package service

import (
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestChannel(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// saltpack.SetLogger(NewLogger(DebugLevel))
	// client.SetLogger(NewLogger(DebugLevel))

	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	ctx := context.TODO()
	testImportKey(t, aliceService, alice)
	testUserSetupGithub(t, env, aliceService, alice, "alice")

	// Alice creates a channel
	channelCreate, err := aliceService.ChannelCreate(ctx, &ChannelCreateRequest{
		Name: "Test",
		User: alice.ID().String(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, channelCreate.Channel)
	channel := channelCreate.Channel

	// Channels (alice)
	channels, err := aliceService.Channels(ctx, &ChannelsRequest{
		User: alice.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(channels.Channels))
	require.Equal(t, "Test", channels.Channels[0].Name)

	export, err := aliceService.KeyExport(ctx, &KeyExportRequest{
		KID:        channel.ID,
		NoPassword: true,
	})
	require.NoError(t, err)

	// Bob service
	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)
	testUserSetupGithub(t, env, bobService, bob, "bob")
	testPull(t, bobService, alice.ID())

	// Channels (bob)
	_, err = bobService.KeyImport(ctx, &KeyImportRequest{
		In: export.Export,
	})
	require.NoError(t, err)
	channels, err = bobService.Channels(ctx, &ChannelsRequest{User: bob.ID().String()})
	require.NoError(t, err)
	require.Equal(t, 1, len(channels.Channels))
	require.Equal(t, channel.ID, channels.Channels[0].ID)
	require.Equal(t, "Test", channels.Channels[0].Name)

	// ChannelCreate (alice@github)
	channelCreate, err = aliceService.ChannelCreate(ctx, &ChannelCreateRequest{
		Name: "Test2",
		User: "alice@github",
	})
	require.NoError(t, err)
	require.NotEmpty(t, channelCreate.Channel)

	// ChannelCreate (unknown key)
	randKey := keys.NewEdX25519KeyFromSeed(testSeed(0xaa))
	_, err = aliceService.ChannelCreate(ctx, &ChannelCreateRequest{
		Name: "Test2",
		User: randKey.ID().String(),
	})
	require.EqualError(t, err, "kex1uu6w5mptvftauu34terj4gz6f3y8u66x8spfa5cxmuhsrdtrddvqevhznx not found")
}
