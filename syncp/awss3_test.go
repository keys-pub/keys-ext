package syncp_test

import (
	"os"
	"testing"

	"github.com/keys-pub/keys-ext/syncp"

	"github.com/stretchr/testify/require"
)

func TestAWSS3(t *testing.T) {
	if os.Getenv("TEST_AWSS3") != "1" {
		t.Skip()
	}
	syncp.SetLogger(syncp.NewLogger(syncp.DebugLevel))

	cfg, closeFn := testConfig(t)
	defer closeFn()

	program, err := syncp.NewAWSS3("s3://keys-pub-awss3-test")
	require.NoError(t, err)

	rt := newTestRuntime(t)
	testProgramSync(t, program, cfg, rt)

	// t.Logf(strings.Join(rt.Logs(), "\n"))
}

func TestAWSS3Fixtures(t *testing.T) {
	if os.Getenv("TEST_AWSS3") != "1" {
		t.Skip()
	}
	syncp.SetLogger(syncp.NewLogger(syncp.DebugLevel))

	cfg, closeFn := testConfig(t)
	defer closeFn()

	program, err := syncp.NewAWSS3("s3://keys-pub-awss3-test")
	require.NoError(t, err)

	testFixtures(t, program, cfg)
}

func TestAWSS3Validate(t *testing.T) {
	_, err := syncp.NewGSUtil("test")
	require.EqualError(t, err, "invalid bucket scheme, expected s3://")
}
