package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

const aliceSeed = "win rebuild update term layer transfer gain field prepare unique spider cool present argue grab trend eagle casino peace hockey loop seed desert swear"
const bobSeed = "crane chimney shell unique drink dynamic math pilot letter inflict tattoo curtain primary crystal live return affair husband general cargo chat vintage demand deer"
const groupSeed = "absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic avoid letter advice comic"

type clock struct {
	t    time.Time
	tick time.Duration
}

func newClock() *clock {
	return newClockAt(1234567890000)
}

func (c *clock) setTick(tick time.Duration) {
	c.tick = tick
}

func newClockAt(ts keys.TimeMs) *clock {
	t := keys.TimeFromMillis(ts)
	return &clock{
		t:    t,
		tick: time.Millisecond,
	}
}

// func newClockAtNow() *clock {
// 	return &clock{
// 		t:    time.Now(),
// 		tick: time.Millisecond,
// 	}
// }

func (c *clock) Now() time.Time {
	c.t = c.t.Add(c.tick)
	return c.t
}

type testServer struct {
	Server  *Server
	Handler http.Handler
}

// func testFirestore(t *testing.T) Fire {
// 	opts := []option.ClientOption{option.WithCredentialsFile("credentials.json")}
// 	fs, fsErr := firestore.NewFirestore("firestore://chilltest-3297b", opts...)
// 	require.NoError(t, fsErr)
// 	err := fs.Delete(context.TODO(), "/")
// 	require.NoError(t, err)
// 	return fs
// }

func testFire(t *testing.T, clock *clock) Fire {
	fi := keys.NewMem()
	fi.SetTimeNow(clock.Now)
	return fi
}

func newTestServer(t *testing.T, clock *clock, fs Fire) *testServer {
	mc := NewMemTestCache(clock.Now)
	server := NewServer(fs, mc)
	tasks := NewTestTasks(server)
	server.search.SetNowFn(clock.Now)
	server.SetTasks(tasks)
	server.SetInternalAuth(keys.RandString(32))
	server.SetNowFn(clock.Now)
	server.SetAccessFn(func(c AccessContext, resource AccessResource, action AccessAction) Access {
		return AccessAllow()
	})
	handler := NewHandler(server)
	return &testServer{
		Server:  server,
		Handler: handler,
	}
}

func (s *testServer) Serve(req *http.Request) (int, http.Header, string) {
	rr := httptest.NewRecorder()
	s.Handler.ServeHTTP(rr, req)
	return rr.Code, rr.Header(), rr.Body.String()
}

// type devServer struct {
// 	client *http.Client
// 	t      *testing.T
// }

// func newDevServer(t *testing.T) *devServer {
// 	client := &http.Client{}
// 	return &devServer{
// 		t:      t,
// 		client: client,
// 	}
// }

// func (s devServer) Serve(req *http.Request) (int, http.Header, string) {
// 	url, err := url.Parse("http://localhost:8080" + req.URL.RequestURI())
// 	require.NoError(s.t, err)
// 	req.URL = url
// 	resp, err := s.client.Do(req)
// 	require.NoError(s.t, err)
// 	b, err := ioutil.ReadAll(resp.Body)
// 	require.NoError(s.t, err)
// 	return resp.StatusCode, resp.Header, string(b)
// }

func TestAccess(t *testing.T) {
	clock := newClock()
	fi := testFire(t, clock)
	srv := newTestServer(t, clock, fi)

	alice, err := keys.NewKeyFromSeedPhrase(aliceSeed, false)
	require.NoError(t, err)
	aliceSpk := alice.PublicKey().SignPublicKey()
	aliceID := alice.ID()

	upkCount := 0
	scCount := 0
	srv.Server.SetAccessFn(func(c AccessContext, resource AccessResource, action AccessAction) Access {
		switch resource {
		case UserPublicKeyResource:
			if action == Put {
				upkCount++
				if upkCount%2 == 0 {
					return AccessDenyTooManyRequests("")
				}
			}
		case SigchainResource:
			if action == Put {
				scCount++
				if scCount == 2 {
					return AccessDenyTooManyRequests("sigchain deny test")
				}
			}
		}
		return AccessAllow()
	})

	// PUT /sigchain/:id/:seq (alice, allow)
	aliceSc := keys.NewSigchain(aliceSpk)
	aliceSt, err := keys.GenerateStatement(aliceSc, []byte("testing"), alice.SignKey(), "", clock.Now())
	require.NoError(t, err)
	err = aliceSc.Add(aliceSt)
	require.NoError(t, err)
	aliceStBytes := aliceSt.Bytes()
	req, err := http.NewRequest("PUT", aliceSt.URLPath(), bytes.NewReader(aliceStBytes))
	require.NoError(t, err)
	code, _, body := srv.Serve(req)
	require.Equal(t, http.StatusOK, code)

	// PUT /sigchain/:id/:seq (alice, deny)
	aliceSt2, err := keys.GenerateStatement(aliceSc, []byte("testing"), alice.SignKey(), "", clock.Now())
	require.NoError(t, err)
	err = aliceSc.Add(aliceSt2)
	require.NoError(t, err)
	aliceStBytes2 := aliceSt2.Bytes()
	req, err = http.NewRequest("PUT", aliceSt2.URLPath(), bytes.NewReader(aliceStBytes2))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusTooManyRequests, code)
	require.Equal(t, `{"error":{"code":429,"message":"sigchain deny test"}}`, body)

	bob, err := keys.NewKeyFromSeedPhrase(bobSeed, false)
	require.NoError(t, err)
	bobSpk := bob.PublicKey().SignPublicKey()

	// PUT /sigchain/:id/:seq (bob, allow)
	bobSc := keys.NewSigchain(bobSpk)
	bobSt, err := keys.GenerateStatement(bobSc, []byte("testing"), bob.SignKey(), "", clock.Now())
	require.NoError(t, err)
	bobAddErr := bobSc.Add(bobSt)
	require.NoError(t, bobAddErr)
	bobStBytes := bobSt.Bytes()
	req, err = http.NewRequest("PUT", bobSt.URLPath(), bytes.NewReader(bobStBytes))
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)

	// POST /task/check/:id
	req, err = http.NewRequest("POST", "/task/check/"+aliceID.String(), nil)
	require.NoError(t, err)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusForbidden, code)
	require.Equal(t, `{"error":{"code":403,"message":"no auth token specified"}}`, body)

	// POST /task/check/:id (with auth)
	req, err = http.NewRequest("POST", "/task/check/"+aliceID.String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", srv.Server.internalAuth)
	code, _, body = srv.Serve(req)
	require.Equal(t, http.StatusOK, code)
}
