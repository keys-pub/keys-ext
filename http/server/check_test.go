package server

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/stretchr/testify/require"
)

func TestUserResult(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))

	clock := newClock()
	fi := testFire(t, clock)
	rq := keys.NewMockRequestor()
	users := testUserStore(t, fi, rq, clock)
	srv := newTestServer(t, clock, fi, users)

	alice, err := keys.NewSignKeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	require.NoError(t, err)

	// Alice sign user statement
	st := userMock(t, users, alice, "alice", "github", rq)

	// PUT /sigchain/:id/:seq
	req, err := http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(st.Bytes()))
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
