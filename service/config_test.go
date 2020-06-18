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

	cfg.SetInt("port", 3001)
	cfg.Set("server", "https://server.url")
	cfg.Set("logLevel", "debug")
	cfg.SetBool("disableSymlinkCheck", true)
	err = cfg.Save()
	require.NoError(t, err)

	cfg2, err := NewConfig("KeysTest")
	require.NoError(t, err)
	require.Equal(t, 3001, cfg2.Port())
	require.Equal(t, "https://server.url", cfg2.Server())
	require.Equal(t, DebugLevel, cfg2.LogLevel())
	require.True(t, cfg2.GetBool("disableSymlinkCheck"))
}
