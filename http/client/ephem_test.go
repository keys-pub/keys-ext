package client

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestEphem(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// api.SetLogger(NewLogger(DebugLevel))
	// server.SetContextLogger(NewContextLogger(DebugLevel))

	env := testEnv(t, logger)
	defer env.closeFn()

	ksa := keys.NewMemKeystore()
	aliceClient := testClient(t, env, ksa)
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	err := ksa.SaveEdX25519Key(alice)
	require.NoError(t, err)

	ksb := keys.NewMemKeystore()
	bobClient := testClient(t, env, ksb)
	bob := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x02}, 32)))
	err = ksb.SaveEdX25519Key(bob)
	require.NoError(t, err)

	// Put
	invite, err := aliceClient.PutEphemeral(context.TODO(), alice.ID(), bob.ID(), []byte("hi"), true)
	require.NoError(t, err)
	require.NotEmpty(t, invite.Code)

	// Get invite
	inviteDetails, err := bobClient.GetInvite(context.TODO(), bob.ID(), invite.Code)
	require.NoError(t, err)
	require.Equal(t, bob.ID(), inviteDetails.Recipient)
	require.Equal(t, alice.ID(), inviteDetails.Sender)

	// Get
	out, err := bobClient.GetEphemeral(context.TODO(), bob.ID(), alice.ID())
	require.NoError(t, err)
	require.Equal(t, []byte("hi"), out)

	// Get (again)
	out, err = bobClient.GetEphemeral(context.TODO(), bob.ID(), alice.ID())
	require.NoError(t, err)
	require.Nil(t, out)

	// Put
	_, err = aliceClient.PutEphemeral(context.TODO(), alice.ID(), bob.ID(), []byte("hi2"), false)
	require.NoError(t, err)

	// Delete
	err = aliceClient.DeleteEphemeral(context.TODO(), alice.ID(), bob.ID())
	require.NoError(t, err)

	// Get
	out, err = bobClient.GetEphemeral(context.TODO(), bob.ID(), alice.ID())
	require.NoError(t, err)
	require.Nil(t, out)

}
