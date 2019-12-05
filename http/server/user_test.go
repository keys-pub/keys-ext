package server

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/stretchr/testify/require"
)

func TestUserCheck(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))

	clock := newClock()
	fi := testFire(t, clock)
	srv := newTestServer(t, clock, fi)

	alice, err := keys.NewKeyFromSeedPhrase(aliceSeed, false)
	require.NoError(t, err)
	aliceID := alice.ID()

	// Alice sign user statement
	aliceSc := keys.NewSigchain(alice.PublicKey().SignPublicKey())
	usr, err := keys.NewUser(aliceID, "test", "alice", "test://", 1)
	require.NoError(t, err)
	aliceSt, err := keys.GenerateUserStatement(aliceSc, usr, alice.SignKey(), clock.Now())
	require.NoError(t, err)

	// PUT /sigchain/:id/:seq
	req, err := http.NewRequest("PUT", aliceSt.URLPath(), bytes.NewReader(aliceSt.Bytes()))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "", body)

	// POST /check
	req, err = api.NewRequest("POST", "/check", nil, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
}
