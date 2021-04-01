package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestUserSearch(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServerEnv(t, env)

	// GET /user/search
	req, err := http.NewRequest("GET", "/user/search", nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[]}`, string(body))

	// Alice
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// Alice sign user statement
	st := userMock(t, alice, "alice", "github", env.client, env.clock)
	// PUT alice
	b, err := st.Bytes()
	require.NoError(t, err)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(b))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", string(body))

	// GET /user/search
	req, err = http.NewRequest("GET", "/user/search", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expected := `{
  "users": [
    {
      "id": "alice@github",
      "name": "alice",
      "kid": "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
      "seq": 1,
      "service": "github",
      "url": "https://gist.github.com/alice/6769746875622f61",
      "status": "ok",
      "statement": "BEGIN MESSAGE.\nY24cziOYlfrf5Jo StDTUuPnS6LKuhx awYPRXyhwwlRHUA c0pXakmOZjgjMzs8DCAifCEdAkZfEU VDk6hs3Z2OGTCKq 6Xr2MZHgg4UNRDb Zy2loGoGN3Mvxd4r7FIwpZOJPE1JEq D2gGjkgLByR9CFG 2aCgRgZZwl5UAa4 6bmBzjEOhmsiW0KTDXulMpC51JXgyc 1MliDDv03o9DXy5 mbXLLP0PDrc9ziK lQqXFL3j737xB4VyAwvIctTRYqHeOH X5y51pUcCb2s5VI uU0x96I37FiTbk8 BW9SG90C6ST.\nEND MESSAGE.",
      "verifiedAt": 1234567890004,
      "ts": 1234567890004
    }
  ]
}`
	require.Equal(t, expected, pretty(t, body))

	// GET /user/search?q=alice
	req, err = http.NewRequest("GET", "/user/search?q=alice", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expected = `{
  "users": [
    {
      "id": "alice@github",
      "name": "alice",
      "kid": "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
      "seq": 1,
      "service": "github",
      "url": "https://gist.github.com/alice/6769746875622f61",
      "status": "ok",
      "statement": "BEGIN MESSAGE.\nY24cziOYlfrf5Jo StDTUuPnS6LKuhx awYPRXyhwwlRHUA c0pXakmOZjgjMzs8DCAifCEdAkZfEU VDk6hs3Z2OGTCKq 6Xr2MZHgg4UNRDb Zy2loGoGN3Mvxd4r7FIwpZOJPE1JEq D2gGjkgLByR9CFG 2aCgRgZZwl5UAa4 6bmBzjEOhmsiW0KTDXulMpC51JXgyc 1MliDDv03o9DXy5 mbXLLP0PDrc9ziK lQqXFL3j737xB4VyAwvIctTRYqHeOH X5y51pUcCb2s5VI uU0x96I37FiTbk8 BW9SG90C6ST.\nEND MESSAGE.",
      "verifiedAt": 1234567890004,
      "ts": 1234567890004
    }
  ]
}`
	require.Equal(t, expected, pretty(t, body))

	// GET /user/search?q=kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077
	req, err = http.NewRequest("GET", "/user/search?q=kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expected = `{
  "users": [
    {
      "id": "alice@github",
      "name": "alice",
      "kid": "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
      "seq": 1,
      "service": "github",
      "url": "https://gist.github.com/alice/6769746875622f61",
      "status": "ok",
      "statement": "BEGIN MESSAGE.\nY24cziOYlfrf5Jo StDTUuPnS6LKuhx awYPRXyhwwlRHUA c0pXakmOZjgjMzs8DCAifCEdAkZfEU VDk6hs3Z2OGTCKq 6Xr2MZHgg4UNRDb Zy2loGoGN3Mvxd4r7FIwpZOJPE1JEq D2gGjkgLByR9CFG 2aCgRgZZwl5UAa4 6bmBzjEOhmsiW0KTDXulMpC51JXgyc 1MliDDv03o9DXy5 mbXLLP0PDrc9ziK lQqXFL3j737xB4VyAwvIctTRYqHeOH X5y51pUcCb2s5VI uU0x96I37FiTbk8 BW9SG90C6ST.\nEND MESSAGE.",
      "verifiedAt": 1234567890004,
      "ts": 1234567890004,
      "mf": "kid"
    }
  ]
}`
	require.Equal(t, expected, pretty(t, body))

	// GET /user/search?q=kbx1rvd43h2sag2tvrdp0duse5p82nvhpjd6hpjwhv7q7vqklega8atshec5ws
	req, err = http.NewRequest("GET", "/user/search?q=kbx1rvd43h2sag2tvrdp0duse5p82nvhpjd6hpjwhv7q7vqklega8atshec5ws", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expected = `{
  "users": [
    {
      "id": "alice@github",
      "name": "alice",
      "kid": "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
      "seq": 1,
      "service": "github",
      "url": "https://gist.github.com/alice/6769746875622f61",
      "status": "ok",
      "statement": "BEGIN MESSAGE.\nY24cziOYlfrf5Jo StDTUuPnS6LKuhx awYPRXyhwwlRHUA c0pXakmOZjgjMzs8DCAifCEdAkZfEU VDk6hs3Z2OGTCKq 6Xr2MZHgg4UNRDb Zy2loGoGN3Mvxd4r7FIwpZOJPE1JEq D2gGjkgLByR9CFG 2aCgRgZZwl5UAa4 6bmBzjEOhmsiW0KTDXulMpC51JXgyc 1MliDDv03o9DXy5 mbXLLP0PDrc9ziK lQqXFL3j737xB4VyAwvIctTRYqHeOH X5y51pUcCb2s5VI uU0x96I37FiTbk8 BW9SG90C6ST.\nEND MESSAGE.",
      "verifiedAt": 1234567890004,
      "ts": 1234567890004,
      "mf": "kid"
    }
  ]
}`
	require.Equal(t, expected, pretty(t, body))

	// GET /user/search?q=unknown
	req, err = http.NewRequest("GET", "/user/search?q=unknown", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"users":[]}`, string(body))
}

func pretty(t *testing.T, b []byte) string {
	var pretty bytes.Buffer
	err := json.Indent(&pretty, b, "", "  ")
	require.NoError(t, err)
	return pretty.String()
}

func TestUserGet(t *testing.T) {
	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServerEnv(t, env)

	// Alice
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	// Alice sign user statement
	st := userMock(t, alice, "alice", "github", env.client, env.clock)
	// PUT alice
	b, err := st.Bytes()
	require.NoError(t, err)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(b))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", string(body))

	// GET /user/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077 (alice)
	req, err = http.NewRequest("GET", "/user/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expected := `{
  "user": {
    "id": "alice@github",
    "name": "alice",
    "kid": "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
    "seq": 1,
    "service": "github",
    "url": "https://gist.github.com/alice/6769746875622f61",
    "status": "ok",
    "statement": "BEGIN MESSAGE.\nY24cziOYlfrf5Jo StDTUuPnS6LKuhx awYPRXyhwwlRHUA c0pXakmOZjgjMzs8DCAifCEdAkZfEU VDk6hs3Z2OGTCKq 6Xr2MZHgg4UNRDb Zy2loGoGN3Mvxd4r7FIwpZOJPE1JEq D2gGjkgLByR9CFG 2aCgRgZZwl5UAa4 6bmBzjEOhmsiW0KTDXulMpC51JXgyc 1MliDDv03o9DXy5 mbXLLP0PDrc9ziK lQqXFL3j737xB4VyAwvIctTRYqHeOH X5y51pUcCb2s5VI uU0x96I37FiTbk8 BW9SG90C6ST.\nEND MESSAGE.",
    "verifiedAt": 1234567890004,
    "ts": 1234567890004
  }
}`
	require.Equal(t, expected, pretty(t, body))

	// GET /user/:kid (not found)
	key := keys.GenerateEdX25519Key()
	req, err = http.NewRequest("GET", "/user/"+key.ID().String(), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"user not found"}}`, string(body))

	// GET /user/search?q=alice@github
	req, err = http.NewRequest("GET", "/user/search?q=alice@github", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expected = `{
  "users": [
    {
      "id": "alice@github",
      "name": "alice",
      "kid": "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
      "seq": 1,
      "service": "github",
      "url": "https://gist.github.com/alice/6769746875622f61",
      "status": "ok",
      "statement": "BEGIN MESSAGE.\nY24cziOYlfrf5Jo StDTUuPnS6LKuhx awYPRXyhwwlRHUA c0pXakmOZjgjMzs8DCAifCEdAkZfEU VDk6hs3Z2OGTCKq 6Xr2MZHgg4UNRDb Zy2loGoGN3Mvxd4r7FIwpZOJPE1JEq D2gGjkgLByR9CFG 2aCgRgZZwl5UAa4 6bmBzjEOhmsiW0KTDXulMpC51JXgyc 1MliDDv03o9DXy5 mbXLLP0PDrc9ziK lQqXFL3j737xB4VyAwvIctTRYqHeOH X5y51pUcCb2s5VI uU0x96I37FiTbk8 BW9SG90C6ST.\nEND MESSAGE.",
      "verifiedAt": 1234567890004,
      "ts": 1234567890004
    }
  ]
}`
	require.Equal(t, expected, pretty(t, body))

	// GET /user/:user (bob@github)
	req, err = http.NewRequest("GET", "/user/bob@github", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusNotFound, code)
	require.Equal(t, `{"error":{"code":404,"message":"user not found"}}`, string(body))

	// GET /user/kbx1rvd43h2sag2tvrdp0duse5p82nvhpjd6hpjwhv7q7vqklega8atshec5ws
	req, err = http.NewRequest("GET", "/user/kbx1rvd43h2sag2tvrdp0duse5p82nvhpjd6hpjwhv7q7vqklega8atshec5ws", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expected = `{
  "user": {
    "id": "alice@github",
    "name": "alice",
    "kid": "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
    "seq": 1,
    "service": "github",
    "url": "https://gist.github.com/alice/6769746875622f61",
    "status": "ok",
    "statement": "BEGIN MESSAGE.\nY24cziOYlfrf5Jo StDTUuPnS6LKuhx awYPRXyhwwlRHUA c0pXakmOZjgjMzs8DCAifCEdAkZfEU VDk6hs3Z2OGTCKq 6Xr2MZHgg4UNRDb Zy2loGoGN3Mvxd4r7FIwpZOJPE1JEq D2gGjkgLByR9CFG 2aCgRgZZwl5UAa4 6bmBzjEOhmsiW0KTDXulMpC51JXgyc 1MliDDv03o9DXy5 mbXLLP0PDrc9ziK lQqXFL3j737xB4VyAwvIctTRYqHeOH X5y51pUcCb2s5VI uU0x96I37FiTbk8 BW9SG90C6ST.\nEND MESSAGE.",
    "verifiedAt": 1234567890004,
    "ts": 1234567890004
  }
}`
	require.Equal(t, expected, pretty(t, body))

	// GET /user/kbx1rvd43h2sag2tvrdp0duse5p82nvhpjd6hpjwhv7q7vqklega8atshec5ws
	req, err = http.NewRequest("GET", "/user/kbx1rvd43h2sag2tvrdp0duse5p82nvhpjd6hpjwhv7q7vqklega8atshec5ws", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expected = `{
  "user": {
    "id": "alice@github",
    "name": "alice",
    "kid": "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
    "seq": 1,
    "service": "github",
    "url": "https://gist.github.com/alice/6769746875622f61",
    "status": "ok",
    "statement": "BEGIN MESSAGE.\nY24cziOYlfrf5Jo StDTUuPnS6LKuhx awYPRXyhwwlRHUA c0pXakmOZjgjMzs8DCAifCEdAkZfEU VDk6hs3Z2OGTCKq 6Xr2MZHgg4UNRDb Zy2loGoGN3Mvxd4r7FIwpZOJPE1JEq D2gGjkgLByR9CFG 2aCgRgZZwl5UAa4 6bmBzjEOhmsiW0KTDXulMpC51JXgyc 1MliDDv03o9DXy5 mbXLLP0PDrc9ziK lQqXFL3j737xB4VyAwvIctTRYqHeOH X5y51pUcCb2s5VI uU0x96I37FiTbk8 BW9SG90C6ST.\nEND MESSAGE.",
    "verifiedAt": 1234567890004,
    "ts": 1234567890004
  }
}`
	require.Equal(t, expected, pretty(t, body))
}

func TestUserDuplicate(t *testing.T) {
	// user.SetLogger(user.NewLogger(user.DebugLevel))

	env := newEnv(t)
	// env.logLevel = server.DebugLevel
	srv := newTestServerEnv(t, env)

	// Alice
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))
	alice2 := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x08}, 32)))

	// PUT /sigchain/alice/1
	st := userMock(t, alice, "alice", "github", env.client, env.clock)
	b, err := st.Bytes()
	require.NoError(t, err)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice.ID()), bytes.NewReader(b))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, "{}", string(body))

	// GET /user/search
	req, err = http.NewRequest("GET", "/user/search", nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
	expected := `{
  "users": [
    {
      "id": "alice@github",
      "name": "alice",
      "kid": "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
      "seq": 1,
      "service": "github",
      "url": "https://gist.github.com/alice/6769746875622f61",
      "status": "ok",
      "statement": "BEGIN MESSAGE.\nY24cziOYlfrf5Jo StDTUuPnS6LKuhx awYPRXyhwwlRHUA c0pXakmOZjgjMzs8DCAifCEdAkZfEU VDk6hs3Z2OGTCKq 6Xr2MZHgg4UNRDb Zy2loGoGN3Mvxd4r7FIwpZOJPE1JEq D2gGjkgLByR9CFG 2aCgRgZZwl5UAa4 6bmBzjEOhmsiW0KTDXulMpC51JXgyc 1MliDDv03o9DXy5 mbXLLP0PDrc9ziK lQqXFL3j737xB4VyAwvIctTRYqHeOH X5y51pUcCb2s5VI uU0x96I37FiTbk8 BW9SG90C6ST.\nEND MESSAGE.",
      "verifiedAt": 1234567890004,
      "ts": 1234567890004
    }
  ]
}`
	require.Equal(t, expected, pretty(t, body))

	// PUT /sigchain/alice2/1
	st2 := userMock(t, alice2, "alice", "github", env.client, env.clock)
	b2, err := st2.Bytes()
	require.NoError(t, err)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/sigchain/%s/1", alice2.ID()), bytes.NewReader(b2))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusConflict, code)
	require.Equal(t, `{"error":{"code":409,"message":"user already exists with key kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077, if you removed or revoked the previous statement you may need to wait briefly for search to update"}}`, string(body))
}
