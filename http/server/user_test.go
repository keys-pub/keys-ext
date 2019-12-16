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
	rq := keys.NewMockRequestor()
	uc := keys.NewTestUserContext(rq, clock.Now)
	srv := newTestServer(t, clock, fi, uc)

	alice, err := keys.NewKeyFromSeedPhrase(aliceSeed, false)
	require.NoError(t, err)

	// Alice sign user statement
	st := userMock(t, uc, alice, "alice", "github", clock, rq)

	// PUT /sigchain/:id/:seq
	req, err := http.NewRequest("PUT", st.URLPath(), bytes.NewReader(st.Bytes()))
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
