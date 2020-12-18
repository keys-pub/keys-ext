package server_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTwitter(t *testing.T) {
	env := newEnv(t)
	srv := newTestServer(t, env)

	testdata, err := ioutil.ReadFile("testdata/1222706272849391616.json")
	require.NoError(t, err)
	api := "https://api.twitter.com/2/tweets/1222706272849391616?expansions=author_id"
	env.client.SetResponse(api, testdata)

	req, err := http.NewRequest("POST", "/twitter/check/gabrlh/1222706272849391616", nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	expected := `BEGIN MESSAGE.EqcgDt8RfXvPq9b 4qCV8S3VPKIQKqa N7Rc1YruQQYuVS8 niHzUv7jdykkEPSrKGcJQCNTkNE7uF swPuwfpaZX6TCKq 6Xr2MZHgg6S0Mjg WFMJ1KHxazTuXs4icK3k8SZCR8mVLQ MSVhFeMrvz0qJOm A96zW9RAY6whsLo 5fC8i3fRJjyo9mQJZid8MwBXJl1XDL 5ZOSkLYs6sk6a2g CiGyA2IP.END MESSAGE.`
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)
}
