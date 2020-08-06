package client_test

import (
	"os"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/firestore"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

const testURL = "firestore://chilltest-3297b"

func testFirestore(t *testing.T) *firestore.Firestore {
	opts := []option.ClientOption{option.WithCredentialsFile("credentials.json")}
	fs, err := firestore.New(testURL, opts...)
	require.NoError(t, err)
	return fs
}

func TestVaultFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	// firestore.SetContextLogger(firestore.NewContextLogger(firestore.DebugLevel))
	fs := testFirestore(t)

	clock := tsutil.NewTestClock()
	env := newEnvWithFire(t, fs, clock, nil)
	// env.logLevel = server.DebugLevel

	alice := keys.GenerateEdX25519Key()

	testVaultMax(t, env, alice)
}
