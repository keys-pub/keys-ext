package syncp_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/syncp"

	"github.com/stretchr/testify/require"
)

func TestGSUtil(t *testing.T) {
	if os.Getenv("TEST_GSUTIL") != "1" {
		t.Skip()
	}
	syncp.SetLogger(syncp.NewLogger(syncp.DebugLevel))

	tmpDir, err := ioutil.TempDir("", "TestGSUtil-"+keys.RandFileName())
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	cfg := syncp.Config{
		Dir: tmpDir,
	}

	existing := map[string][]byte{
		"test.txt":  []byte("testdata"),
		"test2.txt": []byte("testdata2"),
	}

	gsutil, err := syncp.NewGSUtil("gs://keys-chill-test")
	require.NoError(t, err)

	rt := syncp.NewRuntime()
	testProgramSync(t, gsutil, cfg, rt, existing)

	// t.Logf(strings.Join(rt.Logs(), "\n"))
}

func TestGSUtilValidate(t *testing.T) {
	_, err := syncp.NewGSUtil("keys-chill-test")
	require.EqualError(t, err, "invalid bucket scheme, expected gs://")
}
