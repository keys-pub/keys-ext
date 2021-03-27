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
	export, err := aliceService.KeyExport(ctx, &KeyExportRequest{
		KID:        channel.ID,
		NoPassword: true,
	})
	require.NoError(t, err)

	// Alice lists messages
	messages, err := aliceService.Messages(ctx, &MessagesRequest{
		Channel: channel.ID,
		Update:  true,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(messages.Messages))
	require.Equal(t, []string{`alice@github set the channel name to test`}, messages.Messages[0].Text)

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
	_, err = bobService.KeyImport(ctx, &KeyImportRequest{
		In: export.Export,
	})
	require.NoError(t, err)
	_, err = bobService.MessageCreate(ctx, &MessageCreateRequest{
		Sender:  bob.ID().String(),
		Channel: channel.ID,
		Text:    "bm1",
	})
	require.NoError(t, err)

	// Alice lists messages
	messages, err = aliceService.Messages(ctx, &MessagesRequest{
		Channel: channel.ID,
		Update:  true,
	})
	require.NoError(t, err)
	require.Equal(t, 4, len(messages.Messages))

	require.Equal(t, "am1", messages.Messages[1].Text[0])
	require.NotNil(t, messages.Messages[1].Sender)
	require.NotNil(t, messages.Messages[1].Sender.User)
	require.Equal(t, "alice", messages.Messages[1].Sender.User.Name)
	require.Equal(t, "am2", messages.Messages[2].Text[0])
	require.Equal(t, "bm1", messages.Messages[3].Text[0])

	_, err = bobService.Pull(ctx, &PullRequest{Key: alice.ID().String()})
	require.NoError(t, err)

	// Bob lists messages
	messages, err = bobService.Messages(ctx, &MessagesRequest{
		Channel: channel.ID,
		Update:  true,
	})
	require.NoError(t, err)
	require.Equal(t, 4, len(messages.Messages))

	require.Equal(t, "am1", messages.Messages[1].Text[0])
	require.NotNil(t, messages.Messages[1].Sender)
	require.NotNil(t, messages.Messages[1].Sender.User)
	require.Equal(t, "alice", messages.Messages[1].Sender.User.Name)
	require.Equal(t, "am2", messages.Messages[2].Text[0])
	require.Equal(t, "bm1", messages.Messages[3].Text[0])
}
