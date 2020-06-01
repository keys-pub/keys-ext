package service

import (
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestKeyImport(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)

	key := keys.GenerateEdX25519Key()
	export, err := keys.EncodeKeyToSaltpack(key, "testpassword")
	require.NoError(t, err)

	// Import
	importResp, err := service.KeyImport(ctx, &KeyImportRequest{
		In:       []byte(export),
		Password: "testpassword",
	})
	require.NoError(t, err)
	require.Equal(t, key.ID().String(), importResp.KID)

	keyResp, err := service.Key(ctx, &KeyRequest{Identity: key.ID().String()})
	require.NoError(t, err)
	require.Equal(t, key.ID().String(), keyResp.Key.ID)

	// Check key
	ks := service.keyStore()
	out, err := ks.EdX25519Key(key.ID())
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, out.ID(), key.ID())

	sks, err := ks.EdX25519Keys()
	require.NoError(t, err)
	require.Equal(t, 1, len(sks))

	// Import again
	_, err = service.KeyImport(ctx, &KeyImportRequest{
		In:       []byte(export),
		Password: "testpassword",
	})
	require.EqualError(t, err, "keyring item already exists")

	// Import (bob, ID)
	importResp, err = service.KeyImport(ctx, &KeyImportRequest{
		In: []byte(bob.ID().String()),
	})
	require.NoError(t, err)
	require.Equal(t, bob.ID().String(), importResp.KID)

	// Import (charlie, ID with whitespace)
	importResp, err = service.KeyImport(ctx, &KeyImportRequest{
		In: []byte(charlie.ID().String() + "\n  "),
	})
	require.NoError(t, err)
	require.Equal(t, charlie.ID().String(), importResp.KID)

	// Import (error)
	_, err = service.KeyImport(ctx, &KeyImportRequest{In: []byte{}})
	require.EqualError(t, err, "unknown key format")
}
