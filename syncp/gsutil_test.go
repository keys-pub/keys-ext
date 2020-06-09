package syncp_test

import (
	"os"
	"testing"

	"github.com/keys-pub/keys-ext/syncp"

	"github.com/stretchr/testify/require"
)

func TestGSUtil(t *testing.T) {
	if os.Getenv("TEST_GSUTIL") != "1" {
		t.Skip()
	}
	syncp.SetLogger(syncp.NewLogger(syncp.DebugLevel))

	cfg, closeFn := testConfig(t)
	defer closeFn()
	program, err := syncp.NewGSUtil("gs://keys-pub-gsutil-test")
	require.NoError(t, err)

	rt := newTestRuntime(t)
	testProgramSync(t, program, cfg, rt)

	// t.Logf(strings.Join(rt.Logs(), "\n"))
}

func TestGSUtilFixtures(t *testing.T) {
	if os.Getenv("TEST_GSUTIL") != "1" {
		t.Skip()
	}
	syncp.SetLogger(syncp.NewLogger(syncp.DebugLevel))

	cfg, closeFn := testConfig(t)
	defer closeFn()

	program, err := syncp.NewGSUtil("gs://keys-pub-gsutil-test")
	require.NoError(t, err)

	testFixtures(t, program, cfg)
}

func TestGSUtilValidate(t *testing.T) {
	_, err := syncp.NewGSUtil("test")
	require.EqualError(t, err, "invalid bucket scheme, expected gs://")
}
