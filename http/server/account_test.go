package server_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/stretchr/testify/require"
)

func TestAccountCreate(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServerEnv(t, env)
	clock := env.clock
	emailer := srv.Emailer

	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	bob := keys.NewEdX25519KeyFromSeed(testSeed(0x02))

	// PUT /account/:aid
	req, err := http.NewJSONRequest("PUT", dstore.Path("account", alice.ID()), &api.AccountCreateRequest{Email: "alice@keys.pub"}, http.WithTimestamp(clock.Now()), http.SignedWith(alice))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	create := api.AccountCreateResponse{}
	err = json.Unmarshal(body, &create)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, alice.ID(), create.KID)
	verifyCode := emailer.SentVerificationEmail("alice@keys.pub")
	require.NotEmpty(t, verifyCode)

	// GET /account/:aid
	req, err = http.NewAuthRequest("GET", dstore.Path("account", alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	account := api.Account{}
	testJSONUnmarshal(t, body, &account)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, alice.ID(), account.KID)
	require.Equal(t, "alice@keys.pub", account.Email)
	require.False(t, account.VerifiedEmail)

	// POST /account/:aid/verifyemail
	req, err = http.NewJSONRequest("POST", dstore.Path("account", alice.ID(), "verifyemail"), &api.AccountVerifyEmailRequest{Code: verifyCode}, http.WithTimestamp(clock.Now()), http.SignedWith(alice))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	account = api.Account{}
	testJSONUnmarshal(t, body, &account)
	require.Equal(t, http.StatusOK, code)
	require.True(t, account.VerifiedEmail)

	// PUT /account/:aid (email already exists)
	req, err = http.NewJSONRequest("PUT", dstore.Path("account", bob.ID()), &api.AccountCreateRequest{Email: "alice@keys.pub"}, http.WithTimestamp(clock.Now()), http.SignedWith(bob))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusConflict, code)
	require.Equal(t, `{"error":{"code":409,"message":"account already exists"}}`, string(body))

	// PUT /account/:aid (kid already exists)
	req, err = http.NewJSONRequest("PUT", dstore.Path("account", alice.ID()), &api.AccountCreateRequest{Email: "charlie@keys.pub"}, http.WithTimestamp(clock.Now()), http.SignedWith(alice))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusConflict, code)
	require.Equal(t, `{"error":{"code":409,"message":"account already exists"}}`, string(body))

	// PUT /account/:aid (invalid email)
	req, err = http.NewJSONRequest("PUT", dstore.Path("account", bob.ID()), &api.AccountCreateRequest{Email: "alice"}, http.WithTimestamp(clock.Now()), http.SignedWith(bob))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	require.Equal(t, `{"error":{"code":400,"message":"invalid email"}}`, string(body))
}

func TestAccountEmailCodeExpired(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServerEnv(t, env)
	clock := env.clock
	emailer := newTestEmailer()
	srv.Server.SetEmailer(emailer)

	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))

	// PUT /account/:aid
	req, err := http.NewJSONRequest("PUT", dstore.Path("account", alice.ID()), &api.AccountCreateRequest{Email: "alice@keys.pub"}, http.WithTimestamp(clock.Now()), http.SignedWith(alice))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	var create api.AccountCreateResponse
	err = json.Unmarshal(body, &create)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, alice.ID(), create.KID)
	verifyCode := emailer.SentVerificationEmail("alice@keys.pub")
	require.NotEmpty(t, verifyCode)

	// Add hour to clock
	clock.Add(time.Hour)

	// POST /account/:aid/verifyemail
	req, err = http.NewJSONRequest("POST", dstore.Path("account", alice.ID(), "verifyemail"), &api.AccountVerifyEmailRequest{Code: verifyCode}, http.WithTimestamp(clock.Now()), http.SignedWith(alice))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	require.Equal(t, `{"error":{"code":400,"message":"expired code"}}`, string(body))
}

func testAccount(t *testing.T, env *env, srv *testServerEnv, key *keys.EdX25519Key, email string) {
	// PUT /account/:aid
	req, err := http.NewJSONRequest("PUT", dstore.Path("account", key.ID()),
		&api.AccountCreateRequest{Email: email}, http.WithTimestamp(env.clock.Now()), http.SignedWith(key))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	create := api.AccountCreateResponse{}
	err = json.Unmarshal(body, &create)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, key.ID(), create.KID)
}

func testVerifyEmail(t *testing.T, env *env, srv *testServerEnv, key *keys.EdX25519Key, email string) {
	verifyCode := srv.Emailer.SentVerificationEmail(email)
	require.NotEmpty(t, verifyCode)

	// POST /account/:aid/verifyemail
	req, err := http.NewJSONRequest("POST", dstore.Path("account", key.ID(), "verifyemail"), &api.AccountVerifyEmailRequest{Code: verifyCode}, http.WithTimestamp(env.clock.Now()), http.SignedWith(key))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	account := api.Account{}
	testJSONUnmarshal(t, body, &account)
	require.Equal(t, http.StatusOK, code)
	require.True(t, account.VerifiedEmail)
}
