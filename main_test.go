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
	cfg, err := service.NewConfig("KeysTest")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(cfg.AppDir())
		require.NoError(t, err)
	}()
	cfg.SetPort(2001)
	build := service.Build{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
	lgi := service.NewLogrusInterceptor(logrus.StandardLogger())
	serveFn, closeFn, serviceErr := service.NewServiceFn(cfg, build, lgi)
	require.NoError(t, serviceErr)

	go func() {
		time.Sleep(time.Second)
		closeFn()
	}()
	serveErr := serveFn()
	require.NoError(t, serveErr)
}
