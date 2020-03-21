package api

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	tm := keys.TimeFromMillis(123456789000)
	nonce := keys.Bytes32(bytes.Repeat([]byte{0x01}, 32))
	urs := "https://keys.pub/message?version=123456789001"
	auth, err := newAuth("POST", urs, tm, nonce, alice)
	require.NoError(t, err)
	require.Equal(t, "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:sDMBYMJT7OPY1S1eP1I5jmpUSLi4QGAdg2UVooPEkHQwcie8EhfCFZeyeR7D71DkJ6vTb1bOXShmqyOqIk7l7h", auth.Header())
	require.Equal(t, "https://keys.pub/message?nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=123456789000&version=123456789001", auth.URL.String())

	req, err := newRequest(context.TODO(), "POST", urs, nil, tm, nonce, alice)
	require.NoError(t, err)
	require.Equal(t, "https://keys.pub/message?nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=123456789000&version=123456789001", req.URL.String())
	require.Equal(t, "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:sDMBYMJT7OPY1S1eP1I5jmpUSLi4QGAdg2UVooPEkHQwcie8EhfCFZeyeR7D71DkJ6vTb1bOXShmqyOqIk7l7h", req.Header.Get("Authorization"))
}
