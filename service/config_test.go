package service

import (
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
	require.Equal(t, 22405, cfg.Port())

	err = cfg.SetInt("port", 3001, true)
	require.NoError(t, err)
	err = cfg.Set("server", "https://server.url", true)
	require.NoError(t, err)
	err = cfg.Set("logLevel", "debug", true)
	require.NoError(t, err)
	err = cfg.Set("keyringType", "mem", true)
	require.NoError(t, err)
	err = cfg.SetBool("disableSymlinkCheck", true, true)
	require.NoError(t, err)

	cfg2, err := NewConfig("KeysTest")
	require.NoError(t, err)
	require.Equal(t, 3001, cfg2.Port())
	require.Equal(t, "https://server.url", cfg2.Server())
	require.Equal(t, DebugLevel, cfg2.LogLevel())
	require.Equal(t, "mem", cfg2.Get("keyringType", ""))
	require.True(t, cfg2.GetBool("disableSymlinkCheck"))
}
