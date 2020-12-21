package client_test

import (
	"context"
	"os"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/user"
	"github.com/keys-pub/keys/users"
	"github.com/stretchr/testify/require"
)

func TestTwitter(t *testing.T) {
	if os.Getenv("TWITTER_BEARER_TOKEN") == "" {
		t.Skip()
	}

	ctx := context.TODO()
	ds := dstore.NewMem()
	scs := keys.NewSigchains(ds)

	kid := keys.ID("kex1e26rq9vrhjzyxhep0c5ly6rudq7m2cexjlkgknl2z4lqf8ga3uasz3s48m")
	client, err := client.New("https://keys.pub")
	require.NoError(t, err)
	resp, err := client.Sigchain(ctx, kid)
	require.NoError(t, err)
	sc, err := resp.Sigchain()
	require.NoError(t, err)
	err = scs.Save(sc)
	require.NoError(t, err)

	usrs := users.New(ds, scs)
	result, err := usrs.Update(ctx, kid)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, user.StatusOK, result.Status)
}
