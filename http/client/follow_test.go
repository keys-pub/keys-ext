package client_test

import (
	"context"
	"testing"

	"github.com/keys-pub/keys-ext/http/server"
	"github.com/stretchr/testify/require"
)

func TestFollow(t *testing.T) {
	env, closeFn := newEnv(t)
	env.logger = server.NewLogger(server.DebugLevel)
	defer closeFn()
	tk := testKeysSeeded()

	aliceClient := newTestClient(t, env)
	bobClient := newTestClient(t, env)
	alice, bob := tk.alice, tk.bob
	ctx := context.TODO()
	aliceToken := "alicetoken"

	// Alice follow bob
	err := aliceClient.Follow(ctx, alice, bob.ID(), aliceToken)
	require.NoError(t, err)

	// Follows
	follows, err := bobClient.Follows(ctx, bob)
	require.NoError(t, err)
	require.Equal(t, 1, len(follows))
	require.Equal(t, alice.ID(), follows[0].Sender)
	require.Equal(t, bob.ID(), follows[0].Recipient)
	require.Equal(t, "alicetoken", follows[0].Token)

	// FollowedBy
	follow, err := bobClient.FollowedBy(ctx, alice.ID(), bob)
	require.NoError(t, err)
	require.Equal(t, alice.ID(), follow.Sender)
	require.Equal(t, bob.ID(), follow.Recipient)
	require.Equal(t, "alicetoken", follow.Token)

	// Unfollow
	err = bobClient.Unfollow(ctx, alice, bob.ID())
	require.NoError(t, err)
	err = aliceClient.Unfollow(ctx, bob, alice.ID())
	require.EqualError(t, err, "follow not found (404)")

	// Follows
	follows, err = bobClient.Follows(ctx, bob)
	require.NoError(t, err)
	require.Equal(t, 0, len(follows))
}
