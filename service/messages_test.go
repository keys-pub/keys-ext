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

	// Alice creates a channel
	channelCreate, err := aliceService.ChannelCreate(ctx, &ChannelCreateRequest{
		Name:   "Test",
		Member: alice.ID().String(),
	})
	require.NoError(t, err)
	require.NotEmpty(t, channelCreate.Channel)
	channel := channelCreate.Channel

	// Alice invites bob
	_, err = aliceService.ChannelInviteCreate(ctx, &ChannelInviteCreateRequest{
		Channel:   channel.ID,
		Recipient: bob.ID().String(),
		Sender:    alice.ID().String(),
	})
	require.NoError(t, err)
	// Bob accepts invite
	_, err = bobService.ChannelInviteAccept(ctx, &ChannelInviteAcceptRequest{
		Channel: channel.ID,
		Member:  bob.ID().String(),
	})
	require.NoError(t, err)

	// Alice lists messages
	messages, err := aliceService.Messages(ctx, &MessagesRequest{
		Channel: channel.ID,
		Member:  alice.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(messages.Messages))

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
		Member:  alice.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, 3, len(messages.Messages))

	require.Equal(t, "am1", string(messages.Messages[0].Content.Data))
	require.NotNil(t, messages.Messages[0].Sender)
	require.NotNil(t, messages.Messages[0].Sender.User)
	require.Equal(t, "alice", messages.Messages[0].Sender.User.Name)
	require.Equal(t, "am2", string(messages.Messages[1].Content.Data))
	require.Equal(t, "bm1", string(messages.Messages[2].Content.Data))

	_, err = bobService.Pull(ctx, &PullRequest{Key: alice.ID().String()})
	require.NoError(t, err)

	// Bob lists messages
	messages, err = bobService.Messages(ctx, &MessagesRequest{
		Channel: channel.ID,
		Member:  bob.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, 3, len(messages.Messages))

	require.Equal(t, "am1", string(messages.Messages[0].Content.Data))
	require.NotNil(t, messages.Messages[0].Sender)
	require.NotNil(t, messages.Messages[0].Sender.User)
	require.Equal(t, "alice", messages.Messages[0].Sender.User.Name)
	require.Equal(t, "am2", string(messages.Messages[1].Content.Data))
	require.Equal(t, "bm1", string(messages.Messages[2].Content.Data))
}
