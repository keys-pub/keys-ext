package api

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	clock := tsutil.NewTestClock()

	tm := clock.Now()
	nonce := keys.Bytes32(bytes.Repeat([]byte{0x01}, 32))
	urs := "https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?idx=123"
	auth, err := newAuth("POST", urs, tm, nonce, alice)
	require.NoError(t, err)
	require.Equal(t, "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:kOCruiSyqpQBmYAYV74DUot0zxyb39vngKTR4x9rXv8V2DDzdQVyp5Sf1pArnC1enY0Eq3Cnxmg9vW3lBFsx3z", auth.Header())
	require.Equal(t, "https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?idx=123&nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=1234567890001", auth.URL.String())

	req, err := newRequest(context.TODO(), "POST", urs, nil, tm, nonce, alice)
	require.NoError(t, err)
	require.Equal(t, "https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?idx=123&nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=1234567890001", req.URL.String())
	require.Equal(t, "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:kOCruiSyqpQBmYAYV74DUot0zxyb39vngKTR4x9rXv8V2DDzdQVyp5Sf1pArnC1enY0Eq3Cnxmg9vW3lBFsx3z", req.Header.Get("Authorization"))

	rds := NewRedisTest(tsutil.NewTestClock())
	_, err = CheckAuthorization(context.TODO(),
		"POST",
		"https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?idx=123&nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=1234567890001",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:kOCruiSyqpQBmYAYV74DUot0zxyb39vngKTR4x9rXv8V2DDzdQVyp5Sf1pArnC1enY0Eq3Cnxmg9vW3lBFsx3z",
		rds, clock.Now())
	require.NoError(t, err)

	// Change method
	rds = NewRedisTest(tsutil.NewTestClock())
	_, err = CheckAuthorization(context.TODO(),
		"GET",
		"https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?idx=123&nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=1234567890001",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:kOCruiSyqpQBmYAYV74DUot0zxyb39vngKTR4x9rXv8V2DDzdQVyp5Sf1pArnC1enY0Eq3Cnxmg9vW3lBFsx3z",
		rds, clock.Now())
	require.EqualError(t, err, "verify failed")

	// Re-order url params
	rds = NewRedisTest(tsutil.NewTestClock())
	_, err = CheckAuthorization(context.TODO(),
		"POST",
		"https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=1234567890001&idx=123",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:kOCruiSyqpQBmYAYV74DUot0zxyb39vngKTR4x9rXv8V2DDzdQVyp5Sf1pArnC1enY0Eq3Cnxmg9vW3lBFsx3z",
		rds, clock.Now())
	require.EqualError(t, err, "verify failed")

	// Different kid
	rds = NewRedisTest(tsutil.NewTestClock())
	_, err = CheckAuthorization(context.TODO(),
		"POST",
		"https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?idx=123&nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=1234567890001",
		"kex16jvh9cc6na54xwpjs3ztlxdsj6q3scl65lwxxj72m6cadewm404qts0jw9",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:kOCruiSyqpQBmYAYV74DUot0zxyb39vngKTR4x9rXv8V2DDzdQVyp5Sf1pArnC1enY0Eq3Cnxmg9vW3lBFsx3z",
		rds, clock.Now())
	require.EqualError(t, err, "invalid kid")
}
