package service

import (
	"os"
	"strings"
	"testing"

	"github.com/keys-pub/keys"
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

func TestAppPath(t *testing.T) {
	appName := "KeysTest-" + keys.RandFileName()
	cfg, err := NewConfig(appName)
	require.NoError(t, err)

	path, err := cfg.AppPath("", false)
	require.NoError(t, err)

	exists, err := pathExists(path)
	require.NoError(t, err)
	require.False(t, exists)

	path, err = cfg.AppPath("", true)
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(path, appName))
	defer func() { _ = os.RemoveAll(path) }()

	exists, err = pathExists(path)
	require.NoError(t, err)
	require.True(t, exists)

}
