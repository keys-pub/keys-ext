package service

import (
	"context"
	fmt "fmt"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

func TestUserService(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice, true)

	_, err := service.UserService(ctx, &UserServiceRequest{
		Service: "github",
	})
	require.NoError(t, err)
	_, err = service.UserService(ctx, &UserServiceRequest{
		Service: "twitter",
	})
	require.NoError(t, err)
}

func TestUserSign(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice, true)

	resp, err := service.UserSign(ctx, &UserSignRequest{
		KID:     alice.ID().String(),
		Service: "github",
		Name:    "alice",
	})
	require.NoError(t, err)
	require.Equal(t, resp.Name, "alice")

	usr, err := keys.VerifyUser(resp.Message, alice.PublicKey().SignPublicKey(), nil)
	require.NoError(t, err)

	require.Equal(t, "alice", usr.Name)
	require.Equal(t, "github", usr.Service)
	require.Equal(t, alice.ID().String(), usr.KID.String())
}

func TestUserAdd(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice, true)

	testUserSetup(t, env, service, alice.ID(), "alice", true)

	sc, err := service.scs.Sigchain(alice.ID())
	require.NoError(t, err)
	require.Equal(t, 2, len(sc.Statements()))

	resp, err := service.Search(context.TODO(), &SearchRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Results))
	require.Equal(t, 1, len(resp.Results[0].Users))
	require.Equal(t, "alice", resp.Results[0].Users[0].Name)

	testUserSetup(t, env, service, alice.ID(), "alice2", true)

	sc2, err := service.scs.Sigchain(alice.ID())
	require.NoError(t, err)
	require.Equal(t, 3, len(sc2.Statements()))

	resp, err = service.Search(context.TODO(), &SearchRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Results))
	require.Equal(t, 2, len(resp.Results[0].Users))
	require.Equal(t, "alice", resp.Results[0].Users[0].Name)
	require.Equal(t, "alice2", resp.Results[0].Users[1].Name)

	// Try to add user for a public key (not owned)
	randSPK := keys.GenerateSignKey()
	randID := randSPK.ID
	randSC := keys.NewSigchain(randSPK.PublicKey)
	err = service.scs.SaveSigchain(randSC)
	require.NoError(t, err)

	_, err = service.UserAdd(ctx, &UserAddRequest{
		KID:     randID.String(),
		Service: "github",
		Name:    "bob",
		URL:     "https://gist.github.com/bob/1",
	})
	require.EqualError(t, err, fmt.Sprintf("key not found %s", randID))

	// Try to add user for a random ID
	randID2 := keys.RandID()
	_, err = service.UserAdd(ctx, &UserAddRequest{
		KID:     randID2.String(),
		Service: "github",
		Name:    "bob",
		URL:     "https://gist.github.com/bob/1",
	})
	require.EqualError(t, err, fmt.Sprintf("key not found %s", randID2))
}

func TestSearchUsers(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice, false)

	for i := 0; i < 3; i++ {
		keyResp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{})
		require.NoError(t, err)
		_, err = service.Push(ctx, &PushRequest{KID: keyResp.KID})
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
	}

	resp, err := service.Search(ctx, &SearchRequest{})
	require.NoError(t, err)
	require.Equal(t, 3, len(resp.Results))
	require.Equal(t, "username0", resp.Results[0].Users[0].Name)
	require.Equal(t, "username1", resp.Results[1].Users[0].Name)
	require.Equal(t, "username2", resp.Results[2].Users[0].Name)
}

func TestUserAddNotPublished(t *testing.T) {
	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()

	testAuthSetup(t, service, alice, false)

	username := "alice"
	resp, err := service.UserSign(context.TODO(), &UserSignRequest{
		KID:     alice.ID().String(),
		Service: "github",
		Name:    username,
	})
	require.NoError(t, err)

	url := fmt.Sprintf("https://gist.github.com/%s/1", username)
	env.req.SetResponse(url, []byte(resp.Message))

	_, err = service.UserAdd(context.TODO(), &UserAddRequest{
		KID:     alice.ID().String(),
		Service: "github",
		Name:    username,
		URL:     url,
		Local:   false,
	})
	require.EqualError(t, err, "key is not published")
}
