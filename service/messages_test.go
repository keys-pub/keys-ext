package service

import (
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestMessageCreateErrors(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	randKey := keys.GenerateEd25519Key()
	_, err := service.MessageCreate(ctx, &MessageCreateRequest{
		KID:    randKey.ID().String(),
		Sender: alice.ID().String(),
		Text:   "test",
	})
	require.EqualError(t, err, "not found "+randKey.ID().String())
}

func TestMessages(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// saltpack.SetLogger(NewLogger(DebugLevel))
	// client.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := newTestEnv(t)

	aliceService, aliceCloseFn := newTestService(t, env)
	defer aliceCloseFn()
	ctx := context.TODO()
	testAuthSetup(t, aliceService)
	testImportKey(t, aliceService, alice)
	testUserSetup(t, env, aliceService, alice, "alice")
	testPush(t, aliceService, alice)
	testImportKey(t, aliceService, group)

	// Bob service
	bobService, bobCloseFn := newTestService(t, env)
	defer bobCloseFn()
	testAuthSetup(t, bobService)
	testImportKey(t, bobService, bob)
	testUserSetup(t, env, bobService, bob, "bob")
	testImportKey(t, bobService, group)

	// Alice lists messages
	messagesResp, err := aliceService.Messages(ctx, &MessagesRequest{
		KID: group.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(messagesResp.Messages))

	// Check prepare
	_, err = aliceService.MessagePrepare(ctx, &MessagePrepareRequest{
		KID:    group.ID().String(),
		Sender: alice.ID().String(),
		Text:   "prepare",
	})
	require.NoError(t, err)

	// Alice sends 2 messages
	_, err = aliceService.MessageCreate(ctx, &MessageCreateRequest{
		KID:    group.ID().String(),
		Sender: alice.ID().String(),
		Text:   "am1",
	})
	require.NoError(t, err)

	_, messageErrA2 := aliceService.MessageCreate(ctx, &MessageCreateRequest{
		KID:    group.ID().String(),
		Sender: alice.ID().String(),
		Text:   "am2",
	})
	require.NoError(t, messageErrA2)

	// Bob sends message
	_, err = bobService.MessageCreate(ctx, &MessageCreateRequest{
		KID:    group.ID().String(),
		Sender: bob.ID().String(),
		Text:   "bm1",
	})
	require.NoError(t, err)

	// Alice lists messages
	messagesResp2, err := aliceService.Messages(ctx, &MessagesRequest{
		KID: group.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, 3, len(messagesResp2.Messages))

	// cols, err := env.fi.Collections(ctx, "")
	// // cols, err := aliceService.db.Collections(ctx, "")
	// require.NoError(t, err)
	// paths, err := keys.CollectionPaths(cols)
	// require.NoError(t, err)
	// t.Logf("cols: %+v", paths)

	require.Equal(t, "am1", messagesResp2.Messages[0].Content.Text)
	require.NotNil(t, messagesResp2.Messages[0].User)
	require.Equal(t, "alice", messagesResp2.Messages[0].User.Name)
	require.Equal(t, "am2", messagesResp2.Messages[1].Content.Text)
	require.Equal(t, "bm1", messagesResp2.Messages[2].Content.Text)

	_, err = bobService.Pull(ctx, &PullRequest{KID: alice.ID().String()})
	require.NoError(t, err)

	// Bob lists messages
	messagesResp3, err := bobService.Messages(ctx, &MessagesRequest{
		KID: group.ID().String(),
	})
	require.NoError(t, err)
	require.Equal(t, 3, len(messagesResp3.Messages))

	require.Equal(t, "am1", messagesResp3.Messages[0].Content.Text)
	require.NotNil(t, messagesResp3.Messages[0].User)
	require.Equal(t, "alice", messagesResp3.Messages[0].User.Name)
	require.Equal(t, "am2", messagesResp3.Messages[1].Content.Text)
	require.Equal(t, "bm1", messagesResp3.Messages[2].Content.Text)
}
