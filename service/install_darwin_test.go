package service

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/keys-pub/keys/env"
	"github.com/stretchr/testify/require"
)

func TestUninstall(t *testing.T) {
	var out bytes.Buffer
	var err error
	cfg, err := NewConfig("KeyTest")
	require.NoError(t, err)
	err = Uninstall(&out, cfg)
	require.NoError(t, err)

	home := env.MustHomeDir()
	expected := fmt.Sprintf(`Removing "%s/Library/Application Support/KeyTest".
Removing "%s/Library/Logs/KeyTest".
Uninstalled "KeyTest".
`, home, home)
	require.Equal(t, expected, out.String())
}
