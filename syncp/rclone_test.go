package syncp_test

import (
	"os"
	"testing"

	"github.com/keys-pub/keys-ext/syncp"

	"github.com/stretchr/testify/require"
)

func TestRClone(t *testing.T) {
	if os.Getenv("TEST_RCLONE") != "1" {
		t.Skip()
	}
	syncp.SetLogger(syncp.NewLogger(syncp.DebugLevel))

	cfg, closeFn := testConfig(t)
	defer closeFn()

	program, err := syncp.NewRClone("gcs://keys-pub-rclone-test")
	require.NoError(t, err)

	rt := newTestRuntime(t)
	testProgramSync(t, program, cfg, rt)

	// t.Logf(strings.Join(rt.Logs(), "\n"))
}

func TestRCloneFixtures(t *testing.T) {
	if os.Getenv("TEST_RCLONE") != "1" {
		t.Skip()
	}

	syncp.SetLogger(syncp.NewLogger(syncp.DebugLevel))

	cfg, closeFn := testConfig(t)
	defer closeFn()

	program, err := syncp.NewRClone("gcs://keys-pub-rclone-test")
	require.NoError(t, err)

	testFixtures(t, program, cfg)
}
