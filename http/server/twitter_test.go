package server_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/keys-pub/keys/http"
	"github.com/stretchr/testify/require"
)

func TestTwitter(t *testing.T) {
	env := newEnv(t)
	srv := newTestServer(t, env)

	testdata, err := ioutil.ReadFile("testdata/1222706272849391616.json")
	require.NoError(t, err)
	api := "https://api.twitter.com/2/tweets/1222706272849391616?expansions=author_id"
	env.client.SetProxy(api, func(ctx context.Context, req *http.Request, headers []http.Header) http.ProxyResponse {
		return http.ProxyResponse{Body: []byte(testdata)}
	})

	req, err := http.NewRequest("POST", "/twitter/kex1e26rq9vrhjzyxhep0c5ly6rudq7m2cexjlkgknl2z4lqf8ga3uasz3s48m/gabrlh/1222706272849391616", nil)
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	expected := "BEGIN MESSAGE.\nEqcgDt8RfXvPq9b 4qCV8S3VPKIQKqa N7Rc1YruQQYuVS8 niHzUv7jdykkEPSrKGcJQCNTkNE7uF swPuwfpaZX6TCKq 6Xr2MZHgg6S0Mjg WFMJ1KHxazTuXs4icK3k8SZCR8mVLQ MSVhFeMrvz0qJOm A96zW9RAY6whsLo 5fC8i3fRJjyo9mQJZid8MwBXJl1XDL 5ZOSkLYs6sk6a2g CiGyA2IP.\nEND MESSAGE."
	require.Equal(t, expected, body)
	require.Equal(t, http.StatusOK, code)
}
