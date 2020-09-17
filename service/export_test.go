package service

import (
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestKeyExport(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)

	genResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{Type: EdX25519})
	require.NoError(t, err)
	kid, err := keys.ParseID(genResp.KID)
	require.NoError(t, err)

	resp, err := service.KeyExport(ctx, &KeyExportRequest{
		KID:      kid.String(),
		Password: "testpassword",
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Export)

	out, err := keys.DecodeSaltpackKey(string(resp.Export), "testpassword", false)
	require.NoError(t, err)
	require.Equal(t, kid, out.ID())
}

func TestKeySSHExport(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)

	genResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{Type: EdX25519})
	require.NoError(t, err)
	kid, err := keys.ParseID(genResp.KID)
	require.NoError(t, err)

	_, err = service.KeyExport(ctx, &KeyExportRequest{
		KID:  kid.String(),
		Type: SSHExport,
	})
	require.EqualError(t, err, "password required for export")

	resp, err := service.KeyExport(ctx, &KeyExportRequest{
		KID:      kid.String(),
		Type:     SSHExport,
		Password: "testpassword",
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Export)

	_, err = keys.ParseSSHKey(resp.Export, nil, false)
	require.EqualError(t, err, "failed to parse ssh key: ssh: this private key is passphrase protected")

	out, err := keys.ParseSSHKey(resp.Export, []byte("testpassword"), false)
	require.NoError(t, err)
	require.Equal(t, kid, out.ID())

	_, err = service.KeyExport(ctx, &KeyExportRequest{
		KID:      kid.String(),
		Type:     SSHExport,
		Password: "testpassword",
		Public:   true,
	})
	require.EqualError(t, err, "password not supported when exporting public key")

	resp, err = service.KeyExport(ctx, &KeyExportRequest{
		KID:    kid.String(),
		Type:   SSHExport,
		Public: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Export)

	pk, err := keys.ParseSSHPublicKey(string(resp.Export))
	require.NoError(t, err)
	require.Equal(t, kid, pk.ID())

	resp, err = service.KeyExport(ctx, &KeyExportRequest{
		KID:        kid.String(),
		Type:       SSHExport,
		NoPassword: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Export)

	out, err = keys.ParseSSHKey(resp.Export, nil, false)
	require.NoError(t, err)
	require.Equal(t, kid, out.ID())
}
