package server

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keysd/http/api"
	"github.com/stretchr/testify/require"
)

const vaultSeed = "fragile side talk since ready depart adapt flight inquiry memory cupboard settle bracket legal custom razor country task stomach broken reunion roof agree chunk"

func TestVault(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))
	// firestore.SetContextLogger(NewContextLogger(DebugLevel))

	clock := newClock()
	fi := testFire(t, clock)
	rq := keys.NewMockRequestor()
	uc := keys.NewTestUserContext(rq, clock.Now)
	srv := newTestServer(t, clock, fi, uc)

	vault, err := keys.NewKeyFromSeedPhrase(vaultSeed, false)
	require.NoError(t, err)

	// GET /vault/:kid
	req, err := api.NewRequest("GET", keys.Path("vault", vault.ID()), nil, clock.Now(), vault)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)

	// PUT /vault/:kid/:id (no body)
	req, err = api.NewRequest("PUT", keys.Path("vault", vault.ID(), keys.RandID()), nil, clock.Now(), vault)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusBadRequest, code)
	expected := `{"error":{"code":400,"message":"missing body"}}`
	require.Equal(t, expected, body)

	// PUT /vault/:kid/:id
	id := "H1zXH53Xt3JJGx51ruhqk1p83q3VFGmUQCunR51fAsSu"
	req, err = api.NewRequest("PUT", keys.Path("vault", vault.ID(), id), bytes.NewReader([]byte("hi")), clock.Now(), vault)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)

	// POST /vault/:kid/:id (invalid method)
	req, err = api.NewRequest("POST", keys.Path("vault", vault.ID(), keys.RandID()), bytes.NewReader([]byte{}), clock.Now(), vault)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusMethodNotAllowed, code)

	// GET /vault/:kid
	req, err = api.NewRequest("GET", keys.Path("vault", vault.ID()), nil, clock.Now(), vault)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expectedVault := `{"kid":"H2hjAjLKUZmDyZS5QJAEvKK7QjNJ6o8zM9488L2TeZXs","items":[{"data":"aGk=","id":"H1zXH53Xt3JJGx51ruhqk1p83q3VFGmUQCunR51fAsSu","path":"/vault/H1zXH53Xt3JJGx51ruhqk1p83q3VFGmUQCunR51fAsSu"}],"version":"1234567890011"}`
	require.Equal(t, expectedVault, body)

	// GET /vault/:kid?version=1234567890012
	req, err = api.NewRequest("GET", keys.Path("vault", vault.ID())+"?version=1234567890012", nil, clock.Now(), vault)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expectedVault = `{"kid":"H2hjAjLKUZmDyZS5QJAEvKK7QjNJ6o8zM9488L2TeZXs","items":[],"version":"1234567890012"}`
	require.Equal(t, expectedVault, body)
}

func TestVaultAuth(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))
	clock := newClock()
	fi := testFire(t, clock)
	rq := keys.NewMockRequestor()
	uc := keys.NewTestUserContext(rq, clock.Now)
	srv := newTestServer(t, clock, fi, uc)

	vault, err := keys.NewKeyFromSeedPhrase(vaultSeed, false)
	require.NoError(t, err)

	// GET /vault/:kid (no auth)
	req, err := http.NewRequest("GET", keys.Path("vault", keys.RandID()), nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusUnauthorized, code)
	require.Equal(t, `{"error":{"code":401,"message":"missing Authorization header"}}`, body)

	// GET /vault/:kid
	req, err = api.NewRequest("GET", keys.Path("vault", vault.ID()), nil, clock.Now(), vault)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)

	// Replay last request
	reqReplay, err := http.NewRequest("GET", req.URL.String(), nil)
	reqReplay.Header.Set("Authorization", req.Header.Get("Authorization"))
	require.NoError(t, err)
	code, _, body = srv.Serve(reqReplay)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"nonce collision"}}`, body)

	// GET /vault/:kid (invalid authorization)
	authHeader := req.Header.Get("Authorization")
	randKey := keys.GenerateKey()
	sig := strings.Split(authHeader, ":")[1]
	req, err = api.NewRequest("GET", keys.Path("vault", randKey.ID()), nil, clock.Now(), randKey)
	require.NoError(t, err)
	req.Header.Set("Authorization", randKey.ID().String()+":"+sig)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"verify failed"}}`, body)
}
