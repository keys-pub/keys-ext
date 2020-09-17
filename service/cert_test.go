package service

import (
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestCertificate(t *testing.T) {
	env, closeFn := newEnv(t, "", "")
	defer closeFn()

	cert, err := loadCertificate(env)
	require.NoError(t, err)
	require.Empty(t, cert)

	certKey, err := keys.GenerateCertificateKey("localhost", true, nil)
	require.NoError(t, err)
	err = saveCertificate(env, certKey.Public())
	require.NoError(t, err)
	defer func() { _ = DeleteCertificate(env) }()

	cert, err = loadCertificate(env)
	require.NoError(t, err)
	require.NotEmpty(t, cert)

	err = DeleteCertificate(env)
	require.NoError(t, err)

	cert, err = loadCertificate(env)
	require.NoError(t, err)
	require.Empty(t, cert)
}
