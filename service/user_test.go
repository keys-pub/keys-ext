package service

import (
	"context"
	fmt "fmt"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/user"
	"github.com/stretchr/testify/require"
)

func TestUserSearch(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testUserSetupGithub(t, env, service, alice, "alice")
	testPush(t, service, alice)

	testImportKey(t, service, bob)
	testUserSetupGithub(t, env, service, bob, "bob")
	testPush(t, service, bob)

	resp, err := service.UserSearch(ctx, &UserSearchRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(resp.Users))

	// Alice
	require.Equal(t, alice.ID().String(), resp.Users[0].KID)
	require.Equal(t, "alice", resp.Users[0].Name)
	// Charlie
	require.Equal(t, bob.ID().String(), resp.Users[1].KID)

	// TODO: Test stale result
}

func TestUser(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)
	testUserSetupGithub(t, env, service, alice, "alice")

	resp, err := service.User(ctx, &UserRequest{
		KID: alice.ID().String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp.User)
	require.Equal(t, alice.ID().String(), resp.User.KID)

	key := keys.GenerateEdX25519Key()
	resp, err = service.User(ctx, &UserRequest{
		KID: key.ID().String(),
	})
	require.NoError(t, err)
	require.Nil(t, resp.User)
}

func TestUserService(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	_, err := service.UserService(ctx, &UserServiceRequest{
		KID:     alice.ID().String(),
		Service: "github",
	})
	require.NoError(t, err)
	_, err = service.UserService(ctx, &UserServiceRequest{
		KID:     alice.ID().String(),
		Service: "twitter",
	})
	require.NoError(t, err)
}

func TestUserSign(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	resp, err := service.UserSign(ctx, &UserSignRequest{
		KID:     alice.ID().String(),
		Service: "github",
		Name:    "alice",
	})
	require.NoError(t, err)
	require.Equal(t, resp.Name, "alice")

	usr, err := user.Verify(resp.Message, alice.ID(), nil)
	require.NoError(t, err)

	require.Equal(t, "alice", usr.Name)
	require.Equal(t, "github", usr.Service)
	require.Equal(t, alice.ID().String(), usr.KID.String())
}

func TestUserAdd(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	// keys.SetLogger(NewLogger(DebugLevel))
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	testUserSetupGithub(t, env, service, alice, "alice")
	testPush(t, service, alice)

	sc, err := service.scs.Sigchain(alice.ID())
	require.NoError(t, err)
	require.Equal(t, 1, len(sc.Statements()))

	resp, err := service.UserSearch(context.TODO(), &UserSearchRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Users))
	require.Equal(t, "alice", resp.Users[0].Name)

	err = userSetupGithub(env, service, alice, "alice2")
	require.EqualError(t, err, "failed to generate user statement: user set in sigchain already")

	sc2, err := service.scs.Sigchain(alice.ID())
	require.NoError(t, err)
	require.Equal(t, 1, len(sc2.Statements()))

	resp, err = service.UserSearch(context.TODO(), &UserSearchRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Users))
	require.Equal(t, "alice", resp.Users[0].Name)

	// Try to add user for a public key (not owned)
	randSPK := keys.GenerateEdX25519Key()
	randID := randSPK.ID()

	_, err = service.UserAdd(ctx, &UserAddRequest{
		KID:     randID.String(),
		Service: "github",
		Name:    "bob",
		URL:     "https://gist.github.com/bob/1",
	})
	require.EqualError(t, err, fmt.Sprintf("not found %s", randID))

	// Invalid scheme
	_, err = service.UserAdd(ctx, &UserAddRequest{
		KID:     alice.String(),
		Service: "github",
		Name:    "bob",
		URL:     "file://gist.github.com/alice/1",
	})
	require.EqualError(t, err, "failed to create user: invalid scheme for url file://gist.github.com/alice/1")
}

func TestUserAddGithub(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	testAuthSetup(t, service)

	testImportKey(t, service, bob)

	resp, err := service.UserSign(context.TODO(), &UserSignRequest{
		KID:     bob.ID().String(),
		Service: "github",
		Name:    "bob",
	})
	require.NoError(t, err)

	url := fmt.Sprintf("https://gist.github.com/bob/1")
	env.req.SetResponse(url, []byte(resp.Message))

	// Bob
	addResp, err := service.UserAdd(context.TODO(), &UserAddRequest{
		KID:     bob.ID().String(),
		Service: "github",
		Name:    "Bob",
		URL:     "https://gist.github.com/Bob/1",
	})
	require.NoError(t, err)

	require.NotEmpty(t, addResp)
	require.NotEmpty(t, addResp.User)
	require.Equal(t, "bob", addResp.User.Name)
	require.Equal(t, "https://gist.github.com/bob/1", addResp.User.URL)
}

func TestUserAddReddit(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	testAuthSetup(t, service)

	testImportKey(t, service, bob)

	resp, err := service.UserSign(context.TODO(), &UserSignRequest{
		KID:     bob.ID().String(),
		Service: "reddit",
		Name:    "bob",
	})
	require.NoError(t, err)

	url := fmt.Sprintf("https://reddit.com/r/keyspubmsgs/comments/123/bob")
	rmsg := mockRedditMessage("bob", resp.Message, "keyspubmsgs")
	env.req.SetResponse(url+".json", []byte(rmsg))

	// Bob, with funky URL input
	_, err = service.UserAdd(context.TODO(), &UserAddRequest{
		KID:     bob.ID().String(),
		Service: "reddit",
		Name:    "bob",
		URL:     "https://old.reddit.com/r/keyspubmsgs/comments/123/bob/?testing=1",
	})
	require.NoError(t, err)

	// "Bob" sign
	resp, err = service.UserSign(context.TODO(), &UserSignRequest{
		KID:     bob.ID().String(),
		Service: "reddit",
		Name:    "Bob",
	})
	require.EqualError(t, err, "name should be lowercase")

	// "Bob" add
	_, err = service.UserAdd(context.TODO(), &UserAddRequest{
		KID:     bob.ID().String(),
		Service: "reddit",
		Name:    "Bob",
		URL:     "https://old.reddit.com/r/keyspubmsgs/comments/123/bob",
	})
	require.EqualError(t, err, "failed to create user: name should be lowercase")
}

func TestSearchUsers(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env, "")
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	for i := 0; i < 3; i++ {
		keyResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{Type: EdX25519})
		require.NoError(t, err)
		username := fmt.Sprintf("username%d", i)
		kid, err := keys.ParseID(keyResp.KID)
		require.NoError(t, err)

		resp, err := service.UserSign(context.TODO(), &UserSignRequest{
			KID:     kid.String(),
			Service: "github",
			Name:    username,
		})
		require.NoError(t, err)

		url := fmt.Sprintf("https://gist.github.com/%s/1", username)
		env.req.SetResponse(url, []byte(resp.Message))

		_, err = service.UserAdd(context.TODO(), &UserAddRequest{
			KID:     kid.String(),
			Service: "github",
			Name:    username,
			URL:     url,
		})
		require.NoError(t, err)

		_, err = service.Push(context.TODO(), &PushRequest{
			Identity: kid.String(),
		})
		require.NoError(t, err)
	}

	resp, err := service.UserSearch(ctx, &UserSearchRequest{})
	require.NoError(t, err)
	require.Equal(t, 3, len(resp.Users))
	require.Equal(t, "username0", resp.Users[0].Name)
	require.Equal(t, "username1", resp.Users[1].Name)
	require.Equal(t, "username2", resp.Users[2].Name)
}
