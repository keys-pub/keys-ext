package service

import (
	"os"
	"strings"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestEnv(t *testing.T) {
	env, err := NewEnv("KeysTest")
	require.NoError(t, err)
	defer func() {
		removeErr := os.RemoveAll(env.AppDir())
		require.NoError(t, removeErr)
	}()
	require.Equal(t, "KeysTest", env.AppName())
	require.Equal(t, 22405, env.Port())

	env.SetInt("port", 3001)
	env.Set("server", "https://server.url")
	env.SetBool("disableSymlinkCheck", true)
	err = env.Save()
	require.NoError(t, err)

	env2, err := NewEnv("KeysTest")
	require.NoError(t, err)
	require.Equal(t, 3001, env2.Port())
	require.Equal(t, "https://server.url", env2.Server())
	require.True(t, env2.GetBool("disableSymlinkCheck"))
}

func TestAppPath(t *testing.T) {
	appName := "KeysTest-" + keys.RandFileName()
	env, err := NewEnv(appName)
	require.NoError(t, err)

	path, err := env.AppPath("", false)
	require.NoError(t, err)

	exists, err := pathExists(path)
	require.NoError(t, err)
	require.False(t, exists)

	path, err = env.AppPath("", true)
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(path, appName))
	defer func() { _ = os.RemoveAll(path) }()

	exists, err = pathExists(path)
	require.NoError(t, err)
	require.True(t, exists)
}
