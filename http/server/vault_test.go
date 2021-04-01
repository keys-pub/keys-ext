package server_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/firestore"
	"github.com/keys-pub/keys-ext/http/api"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v4"
)

func TestVault(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	testVault(t, env, alice)
}

func TestVaultFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	// firestore.SetContextLogger(firestore.NewContextLogger(firestore.DebugLevel))
	env := newEnvWithFire(t, testFirestore(t), tsutil.NewTestClock())
	// env.logLevel = server.DebugLevel

	key := keys.GenerateEdX25519Key()
	testVault(t, env, key)
}

func testVault(t *testing.T, env *env, alice *keys.EdX25519Key) {
	srv := newTestServerEnv(t, env)
	clock := env.clock

	rand := keys.GenerateEdX25519Key()

	// GET /vault/:kid (not found)
	req, err := http.NewAuthRequest("GET", dstore.Path("vault", alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault not found"}}`, string(body))

	// HEAD /vault/:kid (not found)
	req, err = http.NewAuthRequest("HEAD", dstore.Path("vault", alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)

	// POST /vault/:kid
	vault := []*api.Data{
		{Data: []byte("test1")},
	}
	data, err := json.Marshal(vault)
	require.NoError(t, err)
	contentHash := http.ContentHash(data)
	req, err = http.NewAuthRequest("POST", dstore.Path("vault", alice.ID()), bytes.NewReader(data), contentHash, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, string(body))

	// GET /vault/:kid
	req, err = http.NewAuthRequest("GET", dstore.Path("vault", alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp api.VaultResponse
	err = json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	require.Equal(t, int64(1), resp.Index)
	require.Equal(t, 1, len(resp.Vault))
	require.Equal(t, []byte("test1"), resp.Vault[0].Data)

	// HEAD /vault/:kid
	req, err = http.NewAuthRequest("HEAD", dstore.Path("vault", alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, _ = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)

	// GET /vault/:kid?idx=next
	req, err = http.NewAuthRequest("GET", dstore.Path("vault", alice.ID())+"?idx="+strconv.Itoa(int(resp.Index)), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp2 api.VaultResponse
	err = json.Unmarshal([]byte(body), &resp2)
	require.NoError(t, err)
	require.Equal(t, 0, len(resp2.Vault))
	require.Equal(t, resp.Index, resp2.Index)

	// POST /vault/:kid
	vault2 := []*api.Data{
		{Data: []byte("test2")},
		{Data: []byte("test3")},
	}
	data2, err := json.Marshal(vault2)
	require.NoError(t, err)
	contentHash2 := http.ContentHash(data2)
	req, err = http.NewAuthRequest("POST", dstore.Path("vault", alice.ID()), bytes.NewReader(data2), contentHash2, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, string(body))

	// GET /vault/:kid?idx=next
	req, err = http.NewAuthRequest("GET", dstore.Path("vault", alice.ID())+"?idx="+strconv.Itoa(int(resp.Index)), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp3 api.VaultResponse
	err = json.Unmarshal([]byte(body), &resp3)
	require.NoError(t, err)
	require.Equal(t, 2, len(resp3.Vault))
	require.Equal(t, []byte("test2"), resp3.Vault[0].Data)
	require.Equal(t, []byte("test3"), resp3.Vault[1].Data)

	// POST /vault/:kid
	vault3 := []*api.Data{
		{Data: []byte("test4")},
		{Data: []byte("test5")},
		{Data: []byte("test6")},
		{Data: []byte("test7")},
		{Data: []byte("test8")},
		{Data: []byte("test9")},
	}
	data3, err := json.Marshal(vault3)
	require.NoError(t, err)
	contentHash3 := http.ContentHash(data3)
	req, err = http.NewAuthRequest("POST", dstore.Path("vault", alice.ID()), bytes.NewReader(data3), contentHash3, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, string(body))

	// GET /vault/:kid?idx=next
	req, err = http.NewAuthRequest("GET", dstore.Path("vault", alice.ID())+"?idx="+strconv.Itoa(int(resp3.Index)), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp4 api.VaultResponse
	err = json.Unmarshal([]byte(body), &resp4)
	require.NoError(t, err)
	require.Equal(t, 6, len(resp4.Vault))
	require.Equal(t, []byte("test4"), resp4.Vault[0].Data)
	require.Equal(t, []byte("test5"), resp4.Vault[1].Data)
	require.Equal(t, []byte("test6"), resp4.Vault[2].Data)
	require.Equal(t, []byte("test7"), resp4.Vault[3].Data)
	require.Equal(t, []byte("test8"), resp4.Vault[4].Data)
	require.Equal(t, []byte("test9"), resp4.Vault[5].Data)

	// DEL (invalid auth)
	req, err = http.NewAuthRequest("DELETE", dstore.Path("vault", alice.ID()), nil, "", env.clock.Now(), rand)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, string(body))
	require.Equal(t, http.StatusForbidden, code)

	// DEL /vault/:kid
	req, err = http.NewAuthRequest("DELETE", dstore.Path("vault", alice.ID()), nil, "", env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, string(body))

	// DEL /vault/:kid (again)
	req, err = http.NewAuthRequest("DELETE", dstore.Path("vault", alice.ID()), nil, "", env.clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault was deleted"}}`, string(body))

	// GET /vault/:kid
	req, err = http.NewAuthRequest("GET", dstore.Path("vault", alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault was deleted"}}`, string(body))

	// HEAD /vault/:kid
	req, err = http.NewAuthRequest("HEAD", dstore.Path("vault", alice.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault was deleted"}}`, string(body))

	// POST /vault/:kid (deleted)
	vault = []*api.Data{
		{Data: []byte("testdeleted")},
	}
	data, err = json.Marshal(vault)
	require.NoError(t, err)
	contentHash = http.ContentHash(data)
	req, err = http.NewAuthRequest("POST", dstore.Path("vault", alice.ID()), bytes.NewReader(data), contentHash, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault was deleted"}}`, string(body))
}

func TestVaultAuthFirestore(t *testing.T) {
	if os.Getenv("TEST_FIRESTORE") != "1" {
		t.Skip()
	}
	firestore.SetContextLogger(firestore.NewContextLogger(firestore.DebugLevel))
	fs := testFirestore(t)

	clock := tsutil.NewTestClock()
	env := newEnvWithFire(t, fs, clock)
	// env.logLevel = server.DebugLevel

	alice := keys.GenerateEdX25519Key()

	testVaultAuth(t, env, alice)
}

func TestVaultAuth(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))
	key := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	testVaultAuth(t, env, key)
}

func testVaultAuth(t *testing.T, env *env, key *keys.EdX25519Key) {
	srv := newTestServerEnv(t, env)
	clock := env.clock

	randKey := keys.GenerateEdX25519Key()

	// GET /vault/:kid (no auth)
	req, err := http.NewRequest("GET", dstore.Path("vault", key.ID()), nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"missing Authorization header"}}`, string(body))

	// GET /vault/:kid (invalid key)
	req, err = http.NewAuthRequest("GET", dstore.Path("vault", key.ID()), nil, "", clock.Now(), randKey)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, string(body))

	// POST /vault/:kid (invalid key)
	content := []byte(`[{"data":"dGVzdGluZzE="},{"data":"dGVzdGluZzI="}]`)
	contentHash := http.ContentHash(content)
	req, err = http.NewAuthRequest("POST", dstore.Path("vault", key.ID()), bytes.NewReader(content), contentHash, clock.Now(), randKey)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, string(body))

	// GET /vault/:kid
	req, err = http.NewAuthRequest("GET", dstore.Path("vault", key.ID()), nil, "", clock.Now(), key)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault not found"}}`, string(body))

	// Replay last request
	reqReplay, err := http.NewRequest("GET", req.URL.String(), nil)
	reqReplay.Header.Set("Authorization", req.Header.Get("Authorization"))
	require.NoError(t, err)
	code, _, body = srv.Serve(reqReplay)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"nonce collision"}}`, string(body))

	// GET /vault/:kid (invalid authorization)
	authHeader := req.Header.Get("Authorization")
	sig := strings.Split(authHeader, ":")[1]
	req, err = http.NewAuthRequest("GET", dstore.Path("vault", key.ID()), nil, "", clock.Now(), randKey)
	require.NoError(t, err)
	req.Header.Set("Authorization", randKey.ID().String()+":"+sig)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"invalid kid"}}`, string(body))
}

func TestVaultMsgpack(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))

	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	srv := newTestServerEnv(t, env)
	clock := env.clock

	// POST /vault/:kid
	vault := [][]byte{
		[]byte("test1"),
		[]byte("test2"),
	}
	data, err := msgpack.Marshal(vault)
	require.NoError(t, err)
	contentHash := http.ContentHash(data)
	req, err := http.NewAuthRequest("POST", dstore.Path("vault", alice.ID()+".msgpack"), bytes.NewReader(data), contentHash, clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, string(body))

	// GET /vault/:kid
	req, err = http.NewAuthRequest("GET", dstore.Path("vault", alice.ID()+".msgpack"), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp api.VaultResponse
	err = msgpack.Unmarshal(body, &resp)
	require.NoError(t, err)
	require.Equal(t, int64(2), resp.Index)
	require.Equal(t, 2, len(resp.Vault))
	require.Equal(t, []byte("test1"), resp.Vault[0].Data)
	require.Equal(t, []byte("test2"), resp.Vault[1].Data)
}

func TestVaultAccount(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))

	srv := newTestServerEnv(t, env)
	clock := env.clock
	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))

	vault := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// PUT /vault/:kid (as alice, no account)
	req, err := http.NewAuthRequest("PUT", dstore.Path("vault", vault.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"no account"}}`, string(body))

	testAccount(t, env, srv, alice, "alice@keys.pub")

	// // PUT /vault/:kid (as alice, unverified)
	// req, err = http.NewAuthRequest("PUT", dstore.Path("vault", vault.ID()), nil, "", clock.Now(), alice)
	// require.NoError(t, err)
	// code, _, body = srv.Serve(req)
	// require.Equal(t, http.StatusForbidden, code)
	// require.Equal(t, `{"error":{"code":403,"message":"account email is not verified"}}`, string(body))

	// testVerifyEmail(t, env, srv, alice, "alice@keys.pub")

	// PUT /vault/:kid (as alice, ok)
	req, err = http.NewAuthRequest("PUT", dstore.Path("vault", vault.ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var create api.VaultToken
	testJSONUnmarshal(t, []byte(body), &create)
	require.Equal(t, http.StatusOK, code)
	require.NotEmpty(t, create.Token)

	// POST /vault/:kid
	vault1 := [][]byte{
		bytes.Repeat([]byte{0x01}, 1024),
		bytes.Repeat([]byte{0x02}, 1024),
		bytes.Repeat([]byte{0x03}, 1024),
	}
	data1, err := msgpack.Marshal(vault1)
	require.NoError(t, err)
	req, err = http.NewAuthRequest("POST", dstore.Path("vault", alice.ID())+".msgpack", bytes.NewReader(data1), http.ContentHash(data1), clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, string(body))

	// GET /accounts/:kid/vaults
	req, err = http.NewAuthRequest("GET", dstore.Path("account", alice.ID(), "vaults"), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	var resp api.AccountVaultsResponse
	err = json.Unmarshal(body, &resp)
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Vaults))
	require.Equal(t, int64(3072), resp.Vaults[0].Usage)
}

func TestVaultMax(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))

	srv := newTestServerEnv(t, env)
	clock := env.clock
	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	testAccount(t, env, srv, alice, "alice@keys.pub")
	testVerifyEmail(t, env, srv, alice, "alice@keys.pub")

	// Add too many vaults
	for i := 0; i < 500; i++ {
		vault := keys.GenerateEdX25519Key()
		req, err := http.NewAuthRequest("PUT", dstore.Path("vault", vault.ID()), nil, "", clock.Now(), alice)
		require.NoError(t, err)
		code, _, _ := srv.Serve(req)
		require.Equal(t, http.StatusOK, code)
	}

	req, err := http.NewAuthRequest("PUT", dstore.Path("vault", keys.GenerateEdX25519Key().ID()), nil, "", clock.Now(), alice)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"max vaults reached"}}`, string(body))
}

func TestVaultStatus(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	// keys.SetLogger(keys.NewLogger(keys.DebugLevel))
	srv := newTestServerEnv(t, env)
	clock := env.clock

	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	testAccount(t, env, srv, alice, "alice@keys.pub")
	vault := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// PUT /vault/:kid
	req, err := http.NewAuthRequest("PUT", dstore.Path("vault", vault.ID()), nil, "", clock.Now(), vault)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	var create api.VaultToken
	testJSONUnmarshal(t, []byte(body), &create)
	require.Equal(t, http.StatusOK, code)
	require.NotEmpty(t, create.Token)
	require.Equal(t, vault.ID(), create.KID)
	// Replay is ok
	req, err = http.NewAuthRequest("PUT", dstore.Path("vault", vault.ID()), nil, "", clock.Now(), vault)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.NoError(t, err)
	token := api.VaultToken{}
	testJSONUnmarshal(t, []byte(body), &token)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, token, create)

	// GET /vault/:kid/info
	req, err = http.NewAuthRequest("GET", dstore.Path("vault", vault.ID(), "info"), nil, "", clock.Now(), vault)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	token = api.VaultToken{}
	testJSONUnmarshal(t, []byte(body), &token)
	require.Equal(t, vault.ID(), token.KID)
	require.Equal(t, create.Token, token.Token)

	// GET /vault/:kid/info (unknown)
	unknown := keys.GenerateEdX25519Key()
	req, err = http.NewAuthRequest("GET", dstore.Path("vault", unknown.ID(), "info"), nil, "", clock.Now(), unknown)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"vault not found"}}`, string(body))

	// POST /vault/:kid
	req, err = http.NewJSONRequest("POST", dstore.Path("vault", vault.ID()), []*api.Data{{Data: []byte("test1")}}, http.WithTimestamp(clock.Now()), http.SignedWith(vault))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{}`, string(body))

	// POST /vaults/status
	statusReq := api.VaultsStatusRequest{
		Vaults: map[keys.ID]string{vault.ID(): create.Token},
	}
	req, err = http.NewRequest("POST", "/vaults/status", bytes.NewReader(testJSONMarshal(t, statusReq)))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	var statusResp api.VaultsStatusResponse
	testJSONUnmarshal(t, []byte(body), &statusResp)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, 1, len(statusResp.Vaults))
	require.Equal(t, vault.ID(), statusResp.Vaults[0].ID)
	require.Equal(t, int64(1234567890011), statusResp.Vaults[0].Timestamp)
	require.Equal(t, int64(1), statusResp.Vaults[0].Index)
}
