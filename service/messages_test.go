package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessages(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// saltpack.SetLogger(NewLogger(DebugLevel))
	// client.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	testAuthSetup(t, aliceService)
	ctx := context.TODO()
	testImportKey(t, aliceService, alice)
	testUserSetupGithub(t, env, aliceService, alice, "alice")
	testPush(t, aliceService, alice)

	// Bob service
	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)
	testUserSetupGithub(t, env, bobService, bob, "bob")

	testPull(t, aliceService, bob.ID())

	// Alice creates a channel
	channelCreate, err := aliceService.ChannelCreate(ctx, &ChannelCreateRequest{
		Name: "test",
		User: alice.ID().String(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, channelCreate.Channel)
	channel := channelCreate.Channel

	// Alice invites bob
	_, err = aliceService.ChannelInvitesCreate(ctx, &ChannelInvitesCreateRequest{
		Channel:    channel.ID,
		Sender:     alice.ID().String(),
		Recipients: []string{bob.ID().String()},
	})
	require.NoError(t, err)

	// Bob accepts invite
	_, err = bobService.ChannelJoin(ctx, &ChannelJoinRequest{
		Channel: channel.ID,
		User:    bob.ID().String(),
	})
	require.NoError(t, err)

	// Alice lists messages
	messages, err := aliceService.Messages(ctx, &MessagesRequest{
		Channel: channel.ID,
		User:    alice.ID().String(),
		Update:  true,
	})
	require.NoError(t, err)
	require.Equal(t, 3, len(messages.Messages))
	require.Equal(t, []string{`alice@github set the channel name to test`}, messages.Messages[0].Text)
	require.Equal(t, []string{`alice@github invited [kex1syuhwr4g05t4744r23nvxnr7en9cmz53knhr0gja7c84hr7fkw2quf6zcg]`}, messages.Messages[1].Text)
	require.Equal(t, []string{`bob@github joined`}, messages.Messages[2].Text)

	// Prepare
	_, err = aliceService.MessagePrepare(ctx, &MessagePrepareRequest{
		Channel: channel.ID,
		Sender:  alice.ID().String(),
		Text:    "prepare",
	})
	require.NoError(t, err)

	// Alice sends 2 messages
	_, err = aliceService.MessageCreate(ctx, &MessageCreateRequest{
		Channel: channel.ID,
		Sender:  alice.ID().String(),
		Text:    "am1",
	})
	require.NoError(t, err)

	_, err = aliceService.MessageCreate(ctx, &MessageCreateRequest{
		Channel: channel.ID,
		Sender:  alice.ID().String(),
		Text:    "am2",
	})
	require.NoError(t, err)

	// Bob sends message
	_, err = bobService.MessageCreate(ctx, &MessageCreateRequest{
		Sender:  bob.ID().String(),
		Channel: channel.ID,
		Text:    "bm1",
	})
	require.NoError(t, err)

	// Alice lists messages
	messages, err = aliceService.Messages(ctx, &MessagesRequest{
		Channel: channel.ID,
		User:    alice.ID().String(),
		Update:  true,
	})
	require.NoError(t, err)
	require.Equal(t, 6, len(messages.Messages))

	require.Equal(t, "am1", messages.Messages[3].Text[0])
	require.NotNil(t, messages.Messages[3].Sender)
	require.NotNil(t, messages.Messages[3].Sender.User)
	require.Equal(t, "alice", messages.Messages[3].Sender.User.Name)
	require.Equal(t, "am2", messages.Messages[4].Text[0])
	require.Equal(t, "bm1", messages.Messages[5].Text[0])

	_, err = bobService.Pull(ctx, &PullRequest{Key: alice.ID().String()})
	require.NoError(t, err)

	// Bob lists messages
	messages, err = bobService.Messages(ctx, &MessagesRequest{
		Channel: channel.ID,
		User:    bob.ID().String(),
		Update:  true,
	})
	require.NoError(t, err)
	require.Equal(t, 6, len(messages.Messages))

	require.Equal(t, "am1", messages.Messages[3].Text[0])
	require.NotNil(t, messages.Messages[3].Sender)
	require.NotNil(t, messages.Messages[3].Sender.User)
	require.Equal(t, "alice", messages.Messages[3].Sender.User.Name)
	require.Equal(t, "am2", messages.Messages[4].Text[0])
	require.Equal(t, "bm1", messages.Messages[5].Text[0])
}
