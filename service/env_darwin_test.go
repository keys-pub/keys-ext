package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirs(t *testing.T) {
	env, err := NewEnv("KeysTest")
	require.NoError(t, err)
	require.Equal(t, "KeysTest", env.AppName())

	appDir := env.AppDir()
	require.True(t, strings.HasSuffix(appDir, "/Library/Application Support/KeysTest"))
	logsDir := env.LogsDir()
	require.True(t, strings.HasSuffix(logsDir, "/Library/Logs/KeysTest"))
}
