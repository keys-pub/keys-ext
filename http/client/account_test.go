package client_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/keys-pub/vault/auth"
	"github.com/stretchr/testify/require"
)

func TestAccount(t *testing.T) {
	env, closeFn := newEnv(t, server.NoLevel)
	defer closeFn()
	client := newTestClient(t, env)
	ctx := context.TODO()

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	err := client.AccountCreate(ctx, alice, "alice@keys.pub")
	require.NoError(t, err)

	mk := keys.Rand32()
	pw, err := auth.NewPassword("testpassword", mk)
	require.NoError(t, err)

	err = client.AccountAuthSave(ctx, alice, pw)
	require.NoError(t, err)

	out, err := client.AccountAuths(ctx, alice)
	require.NoError(t, err)
	require.Equal(t, 1, len(out))

	// Create conflict
	err = client.AccountCreate(ctx, alice, "alice@keys.pub")
	require.EqualError(t, err, "account already exists (409)")

	// Create empty
	err = client.AccountCreate(ctx, alice, "")
	require.EqualError(t, err, "invalid email (400)")
}
