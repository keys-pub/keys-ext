package service

import (
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestCertificate(t *testing.T) {
	cfg, closeFn := testConfig(t, "", "mem")
	defer closeFn()

	cert, err := loadCertificate(cfg)
	require.NoError(t, err)
	require.Empty(t, cert)

	certKey, err := keys.GenerateCertificateKey("localhost", true, nil)
	require.NoError(t, err)
	err = saveCertificate(cfg, certKey.Public())
	require.NoError(t, err)
	defer func() { _ = deleteCertificate(cfg) }()

	cert, err = loadCertificate(cfg)
	require.NoError(t, err)
	require.NotEmpty(t, cert)

	err = deleteCertificate(cfg)
	require.NoError(t, err)

	cert, err = loadCertificate(cfg)
	require.NoError(t, err)
	require.Empty(t, cert)
}
