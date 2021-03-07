package server_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/server"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/stretchr/testify/require"
)

func TestAccountCreate(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServer(t, env)
	clock := env.clock
	emailer := newTestEmailer()
	srv.Server.SetEmailer(emailer)

	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	bob := keys.NewEdX25519KeyFromSeed(testSeed(0x02))

	// PUT /account/:kid
	req, err := http.NewJSONRequest("PUT", dstore.Path("account", alice.ID()), &server.AccountCreateRequest{Email: "alice@keys.pub"}, http.WithTimestamp(clock.Now()), http.SignedWith(alice))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	create := server.AccountCreateResponse{}
	err = json.Unmarshal(body, &create)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, alice.ID(), create.KID)
	verifyCode := emailer.SentVerificationEmail("alice@keys.pub")
	require.NotEmpty(t, verifyCode)

	// GET /account/:kid
	req, err = http.NewAuthRequest("GET", dstore.Path("account", alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	account := server.Account{}
	testJSONUnmarshal(t, body, &account)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, alice.ID(), account.KID)
	require.Equal(t, "alice@keys.pub", account.Email)
	require.False(t, account.VerifiedEmail)

	// POST /account/:kid/verifyemail
	req, err = http.NewJSONRequest("POST", dstore.Path("account", alice.ID(), "verifyemail"), &server.AccountVerifyEmailRequest{Code: verifyCode}, http.WithTimestamp(clock.Now()), http.SignedWith(alice))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	account = server.Account{}
	testJSONUnmarshal(t, body, &account)
	require.Equal(t, http.StatusOK, code)
	require.True(t, account.VerifiedEmail)

	// PUT /account/:kid (email already exists)
	req, err = http.NewJSONRequest("PUT", dstore.Path("account", bob.ID()), &server.AccountCreateRequest{Email: "alice@keys.pub"}, http.WithTimestamp(clock.Now()), http.SignedWith(bob))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusConflict, code)
	require.Equal(t, `{"error":{"code":409,"message":"account already exists"}}`, string(body))

	// PUT /account/:kid (kid already exists)
	req, err = http.NewJSONRequest("PUT", dstore.Path("account", alice.ID()), &server.AccountCreateRequest{Email: "charlie@keys.pub"}, http.WithTimestamp(clock.Now()), http.SignedWith(alice))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusConflict, code)
	require.Equal(t, `{"error":{"code":409,"message":"account already exists"}}`, string(body))

	// PUT /account/:kid (invalid email)
	req, err = http.NewJSONRequest("PUT", dstore.Path("account", bob.ID()), &server.AccountCreateRequest{Email: "alice"}, http.WithTimestamp(clock.Now()), http.SignedWith(bob))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	require.Equal(t, `{"error":{"code":400,"message":"invalid email"}}`, string(body))
}

func TestAccountEmailCodeExpired(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServer(t, env)
	clock := env.clock
	emailer := newTestEmailer()
	srv.Server.SetEmailer(emailer)

	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))

	// PUT /account/:kid
	req, err := http.NewJSONRequest("PUT", dstore.Path("account", alice.ID()), &server.AccountCreateRequest{Email: "alice@keys.pub"}, http.WithTimestamp(clock.Now()), http.SignedWith(alice))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	var create server.AccountCreateResponse
	err = json.Unmarshal(body, &create)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, alice.ID(), create.KID)
	verifyCode := emailer.SentVerificationEmail("alice@keys.pub")
	require.NotEmpty(t, verifyCode)

	// Add hour to clock
	clock.Add(time.Hour)

	// POST /account/:kid/verifyemail
	req, err = http.NewJSONRequest("POST", dstore.Path("account", alice.ID(), "verifyemail"), &server.AccountVerifyEmailRequest{Code: verifyCode}, http.WithTimestamp(clock.Now()), http.SignedWith(alice))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	require.Equal(t, `{"error":{"code":400,"message":"expired code"}}`, string(body))
}

func TestAccountVaults(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServer(t, env)
	clock := env.clock
	emailer := newTestEmailer()
	srv.Server.SetEmailer(emailer)

	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))

	// PUT /account/:kid
	req, err := http.NewJSONRequest("PUT", dstore.Path("account", alice.ID()), &server.AccountCreateRequest{Email: "alice@keys.pub"}, http.WithTimestamp(clock.Now()), http.SignedWith(alice))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	var create server.AccountCreateResponse
	err = json.Unmarshal(body, &create)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	verifyCode := emailer.SentVerificationEmail("alice@keys.pub")

	// POST /account/:kid/verifyemail
	req, err = http.NewJSONRequest("POST", dstore.Path("account", alice.ID(), "verifyemail"), &server.AccountVerifyEmailRequest{Code: verifyCode}, http.WithTimestamp(clock.Now()), http.SignedWith(alice))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"email":"alice@keys.pub","kid":"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077","verifiedEmail":true}`+"\n", string(body))

	// PUT /account/:kid/vault/:vid
	vault := keys.GenerateEdX25519Key()
	req, err = http.NewAuthRequest("PUT", dstore.Path("account", alice.ID(), "vault", vault.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}\n", string(body))

	// GET /account/:kid/vaults
	req, err = http.NewAuthRequest("GET", dstore.Path("account", alice.ID(), "vaults"), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	resp := server.AccountVaultsResponse{}
	testJSONUnmarshal(t, body, &resp)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, 1, len(resp.Vaults))
	require.Equal(t, vault.ID(), resp.Vaults[0].VID)
}
