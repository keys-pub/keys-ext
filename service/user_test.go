package service

import (
	"context"
	fmt "fmt"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

// TODO: Service tests are slow

func TestUserService(t *testing.T) {
	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice, true, "")

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
	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice, true, "")

	resp, err := service.UserSign(ctx, &UserSignRequest{
		KID:     alice.ID().String(),
		Service: "test",
		Name:    "alice",
	})
	require.NoError(t, err)
	require.Equal(t, resp.Name, "alice")

	usr, err := keys.VerifyUser(resp.Message, alice.key().PublicKey().SignPublicKey(), nil)
	require.NoError(t, err)

	require.Equal(t, "alice", usr.Name)
	require.Equal(t, "test", usr.Service)
	require.Equal(t, alice.ID().String(), usr.KID.String())
}

func TestSearchUsers(t *testing.T) {
	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()
	testUnlock(t, service)

	for i := 0; i < 3; i++ {
		resp, err := service.KeyGenerate(ctx, &KeyGenerateRequest{})
		require.NoError(t, err)
		_, err = service.Push(ctx, &PushRequest{KID: resp.KID})
		require.NoError(t, err)
		_, err = service.UserAdd(ctx, &UserAddRequest{
			KID:     resp.KID,
			Service: "test",
			Name:    fmt.Sprintf("username%d", i),
			URL:     "test://",
		})
		require.NoError(t, err)
	}

	resp, err := service.Search(ctx, &SearchRequest{})
	require.NoError(t, err)
	require.Equal(t, 3, len(resp.Results))
}

func TestUserAdd(t *testing.T) {
	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice, true, "")

	_, err := service.UserAdd(ctx, &UserAddRequest{
		KID:     alice.ID().String(),
		Service: "test",
		Name:    "alice",
		URL:     "test://",
	})
	require.NoError(t, err)

	sc, err := service.scs.Sigchain(alice.ID())
	require.NoError(t, err)
	require.Equal(t, 2, len(sc.Statements()))

	resp, err := service.Search(context.TODO(), &SearchRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Results))
	require.Equal(t, 1, len(resp.Results[0].Users))
	require.Equal(t, "alice", resp.Results[0].Users[0].Name)

	_, err = service.UserAdd(ctx, &UserAddRequest{
		KID:     alice.ID().String(),
		Service: "test",
		Name:    "alice2",
		URL:     "test://",
	})
	require.NoError(t, err)

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
		Service: "test",
		Name:    "bob",
		URL:     "test://",
	})
	require.EqualError(t, err, fmt.Sprintf("key not found %s", randID))

	// Try to add user for a random ID
	randID2 := keys.RandID()
	_, err = service.UserAdd(ctx, &UserAddRequest{
		KID:     randID2.String(),
		Service: "test",
		Name:    "bob",
		URL:     "test://",
	})
	require.EqualError(t, err, fmt.Sprintf("key not found %s", randID2))
}

func TestUserAddGithub(t *testing.T) {
	service, closeFn := testService(t)
	defer closeFn()
	ctx := context.TODO()
	testAuthSetup(t, service, alice, true, "")

	// signResp, err := service.UserSign(ctx, &UserSignRequest{
	// 	KID:      alice.ID().String(),
	// 	Service:  "github",
	// 	Name: "gabriel",
	// })
	// require.NoError(t, err)
	// t.Logf(signResp.Message)

	addResp, err := service.UserAdd(ctx, &UserAddRequest{
		KID:     alice.ID().String(),
		Service: "github",
		Name:    "gabriel",
		URL:     "https://gist.github.com/gabriel/5e326f50d171f08736f7da3c50e3b9ad",
	})
	require.NoError(t, err)
	require.NotNil(t, addResp)

	resp, err := service.Search(ctx, &SearchRequest{Query: alice.ID().String()})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Results))
	require.Equal(t, 1, len(resp.Results[0].Users))
	require.Equal(t, "gabriel", resp.Results[0].Users[0].Name)
	require.Equal(t, "github", resp.Results[0].Users[0].Service)
}
