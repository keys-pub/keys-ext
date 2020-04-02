package server_test

import (
	"context"
	"testing"

	"github.com/keys-pub/keysd/firestore"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

const testURL = "firestore://chilltest-3297b"

func testFirestore(t *testing.T, clear bool) *firestore.Firestore {
	opts := []option.ClientOption{option.WithCredentialsFile("credentials.json")}
	fs, err := firestore.NewFirestore(testURL, opts...)
	require.NoError(t, err)
	fs.Test = true
	require.NoError(t, err)
	if clear {
		_, err := fs.Delete(context.TODO(), "/")
		require.NoError(t, err)
	}
	return fs
}

func TestMessagesFirestore(t *testing.T) {
	t.Skip()
	firestore.SetContextLogger(firestore.NewContextLogger(firestore.DebugLevel))
	fs := testFirestore(t, true)

	clock := newClock()
	env := newEnvWithFire(t, fs, clock)
	// env.logLevel = server.DebugLevel

	testMessages(t, env)
}
