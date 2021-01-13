package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirs(t *testing.T) {
	env, err := NewEnv("KeysTest", build)
	require.NoError(t, err)
	require.Equal(t, "KeysTest", env.AppName())

	appDir := env.AppDir()
	require.True(t, strings.HasSuffix(appDir, `/.local/share/KeysTest`))
	logsDir := env.LogsDir()
	require.True(t, strings.HasSuffix(logsDir, `/.cache/KeysTest`))
}
