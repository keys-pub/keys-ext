package service

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	cfg, err := NewConfig("KeysTest")
	require.NoError(t, err)
	defer func() {
		removeErr := os.RemoveAll(cfg.AppDir())
		require.NoError(t, removeErr)
	}()
	require.Equal(t, "KeysTest", cfg.AppName())
	require.Equal(t, 10001, cfg.Port())

	cfg.SetPort(3001)
	err = cfg.Save()
	require.NoError(t, err)

	cfg2, err := NewConfig("KeysTest")
	require.NoError(t, err)
	require.Equal(t, 3001, cfg2.Port())
}

func TestConfigSet(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service, alice)

	_, err := service.ConfigSet(ctx, &ConfigSetRequest{
		Key:   "disablePromptUser",
		Value: "??",
	})
	require.EqualError(t, err, "invalid value: ??")

	_, err = service.ConfigSet(ctx, &ConfigSetRequest{
		Key:   "unknown",
		Value: "1",
	})
	require.EqualError(t, err, "unknown config key")

	_, err = service.ConfigSet(ctx, &ConfigSetRequest{
		Key:   "disablePromptUser",
		Value: "1",
	})

	configResp, configErr := service.Config(ctx, &ConfigRequest{})
	require.NoError(t, configErr)
	require.Equal(t, "1", configResp.Config["disablePromptUser"])
}
