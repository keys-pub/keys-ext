package service

import (
	strings "strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigDirs(t *testing.T) {
	cfg, err := NewConfig("KeysTest")
	require.NoError(t, err)
	require.Equal(t, "KeysTest", cfg.AppName())

	appDir := cfg.AppDir()
	require.True(t, strings.HasSuffix(appDir, "/Library/Application Support/KeysTest"))
	logsDir := cfg.LogsDir()
	require.True(t, strings.HasSuffix(logsDir, "/Library/Logs/KeysTest"))
}
