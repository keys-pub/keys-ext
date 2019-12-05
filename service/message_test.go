package service

import (
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestMessageCreateErrors(t *testing.T) {
	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice, true, "")

	randID := keys.RandID()
	_, messageErr := service.MessageCreate(ctx, &MessageCreateRequest{
		KID:    randID.String(),
		Sender: alice.ID().String(),
		Text:   "test",
	})
	require.EqualError(t, messageErr, "key not found "+randID.String())
}

func TestMessages(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// saltpack.SetLogger(NewLogger(DebugLevel))
	// client.SetLogger(NewLogger(DebugLevel))

	clock := newClock()
	fi := testFire(t, clock)

	aliceService, aliceCloseFn := testServiceFire(t, fi, clock)
	defer aliceCloseFn()
	ctx := context.TODO()
	testAuthSetup(t, aliceService, alice, true, "alice")

	// Bob service
	bobService, bobCloseFn := testServiceFire(t, fi, clock)
	defer bobCloseFn()
	testAuthSetup(t, bobService, bob, true, "bob")

	group := keys.GenerateKey()
	errG1 := aliceService.ks.SaveKey(group, true, clock.Now())
	require.NoError(t, errG1)
	errG2 := bobService.ks.SaveKey(group, true, clock.Now())
	require.NoError(t, errG2)

	// Alice lists messages
	messagesResp, messagesErr := aliceService.Messages(ctx, &MessagesRequest{
		KID: group.ID().String(),
	})
	require.NoError(t, messagesErr)
	require.Equal(t, 0, len(messagesResp.Messages))

	// Check prepare
	_, messageErrAP1 := aliceService.MessagePrepare(ctx, &MessagePrepareRequest{
		KID:    group.ID().String(),
		Sender: alice.ID().String(),
		Text:   "prepare",
	})
	require.NoError(t, messageErrAP1)

	// Alice sends 2 messages
	_, messageErrA1 := aliceService.MessageCreate(ctx, &MessageCreateRequest{
		KID:    group.ID().String(),
		Sender: alice.ID().String(),
		Text:   "am1",
	})
	require.NoError(t, messageErrA1)

	_, messageErrA2 := aliceService.MessageCreate(ctx, &MessageCreateRequest{
		KID:    group.ID().String(),
		Sender: alice.ID().String(),
		Text:   "am2",
	})
	require.NoError(t, messageErrA2)

	// Bob sends message
	_, messageErrB1 := bobService.MessageCreate(ctx, &MessageCreateRequest{
		KID:    group.ID().String(),
		Sender: bob.ID().String(),
		Text:   "bm1",
	})
	require.NoError(t, messageErrB1)

	// Alice lists messages
	messagesResp2, messagesErr2 := aliceService.Messages(ctx, &MessagesRequest{
		KID: group.ID().String(),
	})
	require.NoError(t, messagesErr2)
	require.Equal(t, 3, len(messagesResp2.Messages))

	// iter, iterErr := aliceService.db.Iterator(context.TODO(), "", nil)
	// require.NoError(t, iterErr)
	// defer iter.Release()
	// spew, err := keys.Spew(iter, nil)
	// require.NoError(t, err)
	// t.Logf(spew.String())

	require.Equal(t, "am1", messagesResp2.Messages[0].Content.Text)
	require.Equal(t, "alice", messagesResp2.Messages[0].User.Name)
	require.Equal(t, "am2", messagesResp2.Messages[1].Content.Text)
	require.Equal(t, "bm1", messagesResp2.Messages[2].Content.Text)

	// // Alice lists messages (index=2)
	// messagesResp2, messagesErr2 := aliceService.Messages(ctx, &MessagesRequest{
	// 	KID: group.ID().String(),
	// })
	// require.NoError(t, messagesErr2)
	// require.Equal(t, 1, len(messagesResp2.Messages))
	// require.Equal(t, "bm1", messagesResp2.Messages[0].Content.Text)

	_, err := bobService.Pull(ctx, &PullRequest{KID: alice.ID().String()})
	require.NoError(t, err)

	// Bob lists messages
	messagesResp3, messagesErr3 := bobService.Messages(ctx, &MessagesRequest{
		KID: group.ID().String(),
	})
	require.NoError(t, messagesErr3)
	require.Equal(t, 3, len(messagesResp3.Messages))

	require.Equal(t, "am1", messagesResp3.Messages[0].Content.Text)
	require.NotNil(t, messagesResp3.Messages[0].User)
	require.Equal(t, "alice", messagesResp3.Messages[0].User.Name)
	require.Equal(t, "am2", messagesResp3.Messages[1].Content.Text)
	require.Equal(t, "bm1", messagesResp3.Messages[2].Content.Text)
}
