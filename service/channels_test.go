package service

// func TestChannel(t *testing.T) {
// 	// SetLogger(NewLogger(DebugLevel))
// 	// saltpack.SetLogger(NewLogger(DebugLevel))
// 	// client.SetLogger(NewLogger(DebugLevel))

// 	env := newTestEnv(t)

// 	aliceService, aliceCloseFn := newTestService(t, env)
// 	defer aliceCloseFn()
// 	testAuthSetup(t, aliceService)
// 	ctx := context.TODO()
// 	testImportKey(t, aliceService, alice)
// 	testUserSetupGithub(t, env, aliceService, alice, "alice")
// 	testPush(t, aliceService, alice)

// 	// Bob service
// 	bobService, bobCloseFn := newTestService(t, env)
// 	defer bobCloseFn()
// 	testAuthSetup(t, bobService)
// 	testImportKey(t, bobService, bob)
// 	testUserSetupGithub(t, env, bobService, bob, "bob")
// 	testPush(t, bobService, bob)
// 	testPull(t, bobService, alice.ID())

// 	// Alice creates a channel
// 	channelCreate, err := aliceService.ChannelCreate(ctx, &ChannelCreateRequest{
// 		Name: "Test",
// 		User: alice.ID().String(),
// 	})
// 	require.NoError(t, err)
// 	require.NotEmpty(t, channelCreate.Channel)
// 	cid := channelCreate.Channel.ID

// 	// Channels (alice)
// 	channels, err := aliceService.Channels(ctx, &ChannelsRequest{
// 		User: alice.ID().String(),
// 	})
// 	require.NoError(t, err)
// 	require.Equal(t, 1, len(channels.Channels))
// 	require.Equal(t, "Test", channels.Channels[0].Name)

// 	// Channels (alice@github)
// 	channels, err = aliceService.Channels(ctx, &ChannelsRequest{
// 		User: "alice@github",
// 	})
// 	require.NoError(t, err)
// 	require.Equal(t, 1, len(channels.Channels))
// 	require.Equal(t, "Test", channels.Channels[0].Name)

// 	// Alice invites bob
// 	_, err = aliceService.ChannelInvite(ctx, &ChannelInviteRequest{
// 		Channel:    cid,
// 		Sender:     alice.ID().String(),
// 		Recipients: []string{bob.ID().String()},
// 	})
// 	require.NoError(t, err)

// 	// Bob joins (accepts invite)
// 	_, err = bobService.ChannelJoin(ctx, &ChannelJoinRequest{
// 		Channel: cid,
// 		User:    bob.ID().String(),
// 	})
// 	require.NoError(t, err)

// 	// Channels (bob)
// 	channels, err = bobService.Channels(ctx, &ChannelsRequest{
// 		User: bob.ID().String(),
// 	})
// 	require.NoError(t, err)
// 	require.Equal(t, 1, len(channels.Channels))
// 	require.Equal(t, "Test", channels.Channels[0].Name)

// 	// ChannelCreate (alice@github)
// 	channelCreate, err = aliceService.ChannelCreate(ctx, &ChannelCreateRequest{
// 		Name: "Test2",
// 		User: "alice@github",
// 	})
// 	require.NoError(t, err)
// 	require.NotEmpty(t, channelCreate.Channel)

// 	// ChannelCreate (unknown key)
// 	randKey := keys.NewEdX25519KeyFromSeed(testSeed(0xaa))
// 	_, err = aliceService.ChannelCreate(ctx, &ChannelCreateRequest{
// 		Name: "Test2",
// 		User: randKey.ID().String(),
// 	})
// 	require.EqualError(t, err, "kex1uu6w5mptvftauu34terj4gz6f3y8u66x8spfa5cxmuhsrdtrddvqevhznx not found")

// 	// Channels (unknown key)
// 	_, err = aliceService.Channels(ctx, &ChannelsRequest{
// 		User: randKey.ID().String(),
// 	})
// 	require.EqualError(t, err, "kex1uu6w5mptvftauu34terj4gz6f3y8u66x8spfa5cxmuhsrdtrddvqevhznx not found")

// 	// Channels (unknown user)
// 	_, err = aliceService.Channels(ctx, &ChannelsRequest{
// 		User: "unknown@github",
// 	})
// 	require.EqualError(t, err, "unknown@github not found")

// 	// Channels (unauthorized)
// 	_, err = bobService.Channels(ctx, &ChannelsRequest{
// 		User: "alice@github",
// 	})
// 	require.EqualError(t, err, "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077 not found")
// }
