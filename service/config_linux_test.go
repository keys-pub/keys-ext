package service

import (
	strings "strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigDirs(t *testing.T) {
	cfg, err := NewConfig("KeysTest", 2000)
	require.NoError(t, err)
	require.Equal(t, "KeysTest", cfg.AppName())

	appDir := cfg.AppDir()
	require.True(t, strings.HasSuffix(appDir, `/.local/share/KeysTest`))
	logsDir := cfg.LogsDir()
	require.True(t, strings.HasSuffix(logsDir, `/.cache/KeysTest`))
}
