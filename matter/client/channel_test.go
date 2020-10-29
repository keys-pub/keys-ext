package client_test

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys-ext/matter/client"
	"github.com/stretchr/testify/require"
)

func TestCreateChannel(t *testing.T) {
	var err error
	ctx := context.TODO()

	cl, err := client.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	user, err := cl.LoginWithPassword(ctx, "testuser", "testuserpassword")
	require.NoError(t, err)
	require.NotNil(t, user)

	team, err := cl.TeamByName(context.TODO(), "test")
	require.NoError(t, err)
	require.NotNil(t, team)

	channel, err := cl.CreateChannel(ctx, &client.Channel{
		Name:   "test2",
		TeamID: team.ID,
		Type:   client.ChannelPrivate,
	})
	require.NoError(t, err)
	require.NotNil(t, channel)

	channels, err := cl.ChannelsForUser(ctx, "", team.ID)
	require.NoError(t, err)
	require.NotEmpty(t, channels)
	spew.Dump(channels)
}
