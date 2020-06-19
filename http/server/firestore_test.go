package server_test

import (
	"os"
	"testing"

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

func TestMessagesFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	firestore.SetContextLogger(firestore.NewContextLogger(firestore.DebugLevel))
	fs := testFirestore(t)

	clock := tsutil.NewClock()
	env := newEnvWithFire(t, fs, clock)
	// env.logLevel = server.DebugLevel

	testMessages(t, env)
}
