package main

import (
	"os"
	"testing"
	"time"

	"github.com/keys-pub/keysd/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestServiceFn(t *testing.T) {
	// service.SetLogger(service.NewLogger(service.DebugLevel))
	cfg, err := service.NewConfig("KeysServiceTest")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(cfg.AppDir())
		require.NoError(t, err)
	}()
	cfg.SetInt("port", 2001)
	cfg.Set("keyring", "mem")
	build := service.Build{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
	lgi := service.NewLogrusInterceptor(logrus.StandardLogger())
	cert, err := service.GenerateCertificate(cfg, false)
	require.NoError(t, err)
	serveFn, closeFn, serviceErr := service.NewServiceFn(cfg, build, cert, lgi)
	require.NoError(t, serviceErr)

	t.Logf("testing")
	go func() {
		time.Sleep(time.Second)
		closeFn()
	}()
	serveErr := serveFn()
	require.NoError(t, serveErr)
}
