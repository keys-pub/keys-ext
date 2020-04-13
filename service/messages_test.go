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

	// Alice lists messages
	messagesResp, err := aliceService.Messages(ctx, &MessagesRequest{
		Sender:    alice.ID().String(),
		Recipient: bob.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(messagesResp.Messages))

	// Check prepare
	_, err = aliceService.MessagePrepare(ctx, &MessagePrepareRequest{
		Sender:    alice.ID().String(),
		Recipient: bob.ID().String(),
		Text:      "prepare",
	})
	require.NoError(t, err)

	// Alice sends 2 messages
	_, err = aliceService.MessageCreate(ctx, &MessageCreateRequest{
		Sender:    alice.ID().String(),
		Recipient: bob.ID().String(),
		Text:      "am1",
	})
	require.NoError(t, err)

	_, messageErrA2 := aliceService.MessageCreate(ctx, &MessageCreateRequest{
		Sender:    alice.ID().String(),
		Recipient: bob.ID().String(),
		Text:      "am2",
	})
	require.NoError(t, messageErrA2)

	// Bob sends message
	_, err = bobService.MessageCreate(ctx, &MessageCreateRequest{
		Sender:    bob.ID().String(),
		Recipient: alice.ID().String(),
		Text:      "bm1",
	})
	require.NoError(t, err)

	// Alice lists messages
	messagesResp2, err := aliceService.Messages(ctx, &MessagesRequest{
		Sender:    alice.ID().String(),
		Recipient: bob.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, 3, len(messagesResp2.Messages))

	// cols, err := env.fi.Collections(ctx, "")
	// // cols, err := aliceService.db.Collections(ctx, "")
	// require.NoError(t, err)
	// paths, err := ds.CollectionPaths(cols)
	// require.NoError(t, err)
	// t.Logf("cols: %+v", paths)

	require.Equal(t, "am1", string(messagesResp2.Messages[0].Content.Data))
	require.NotNil(t, messagesResp2.Messages[0].Sender)
	require.NotNil(t, messagesResp2.Messages[0].Sender.User)
	require.Equal(t, "alice", messagesResp2.Messages[0].Sender.User.Name)
	require.Equal(t, "am2", string(messagesResp2.Messages[1].Content.Data))
	require.Equal(t, "bm1", string(messagesResp2.Messages[2].Content.Data))

	_, err = bobService.Pull(ctx, &PullRequest{Identity: alice.ID().String()})
	require.NoError(t, err)

	// Bob lists messages
	messagesResp3, err := bobService.Messages(ctx, &MessagesRequest{
		Sender:    bob.ID().String(),
		Recipient: alice.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, 3, len(messagesResp3.Messages))

	require.Equal(t, "am1", string(messagesResp3.Messages[0].Content.Data))
	require.NotNil(t, messagesResp3.Messages[0].Sender)
	require.NotNil(t, messagesResp3.Messages[0].Sender.User)
	require.Equal(t, "alice", messagesResp3.Messages[0].Sender.User.Name)
	require.Equal(t, "am2", string(messagesResp3.Messages[1].Content.Data))
	require.Equal(t, "bm1", string(messagesResp3.Messages[2].Content.Data))
}
