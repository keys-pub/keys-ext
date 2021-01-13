package main

import (
	"os"
	"testing"
	"time"

	"github.com/keys-pub/keys-ext/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestServiceFn(t *testing.T) {
	build := service.Build{
		Version: version,
		Commit:  commit,
		Date:    date,
	}

	// service.SetLogger(service.NewLogger(service.DebugLevel))
	env, err := service.NewEnv("KeysServiceTest", build)
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(env.AppDir())
		require.NoError(t, err)
	}()
	env.SetInt("port", 2001)
	lgi := service.NewLogrusInterceptor(logrus.StandardLogger())
	cert, err := service.GenerateCertificate(env, false)
	require.NoError(t, err)
	serveFn, closeFn, serviceErr := service.NewServiceFn(env, build, cert, lgi)
	require.NoError(t, serviceErr)

	t.Logf("testing")
	go func() {
		time.Sleep(time.Second)
		closeFn()
	}()
	serveErr := serveFn()
	require.NoError(t, serveErr)

	// Give time to close
	time.Sleep(time.Second)
}
