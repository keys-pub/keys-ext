package wormhole_test

import (
	"testing"

	"github.com/keys-pub/keys"

	"github.com/keys-pub/keysd/wormhole"
	"github.com/stretchr/testify/require"
)

func TestWormhole(t *testing.T) {
	alice := keys.GenerateX25519Key()
	bob := keys.GenerateX25519Key()

	aliceWh, err := wormhole.NewWormhole(alice, bob.PublicKey(), true)
	require.NoError(t, err)

	bobWh, err := wormhole.NewWormhole(bob, alice.PublicKey(), false)
	require.NoError(t, err)

	aliceWh.Start()
}
