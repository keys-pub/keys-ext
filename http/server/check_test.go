package server_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/http"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))

	env := newEnv(t)
	srv := newTestServerEnv(t, env)
	clock := env.clock

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// Alice sign user statement
	st := userMock(t, alice, "alice", "github", env.client, clock)

	// PUT /sigchain/:id/:seq
	b, err := st.Bytes()
	require.NoError(t, err)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(b))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", string(string(body)))

	// POST /check
	req, err = http.NewAuthRequest("POST", "/check", nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", string(body))
}
