package client_test

import (
	"context"
	"testing"

	"github.com/keys-pub/keys-ext/matter/client"
	"github.com/stretchr/testify/require"
)

func TestAdminUser(t *testing.T) {
	var err error
	ctx := context.TODO()

	cl, err := client.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	admin, err := cl.LoginWithPassword(ctx, "gabriel", "testpassword")
	require.NoError(t, err)

	team, err := cl.TeamByName(ctx, "test")
	require.NoError(t, err)
	require.NotNil(t, team)

	err = cl.AddUserToTeam(ctx, admin.ID, team.ID)
	require.NoError(t, err)

	channel, err := cl.ChannelByName(ctx, team.ID, "town-square")
	require.NoError(t, err)

	err = cl.AddUserToChannel(ctx, admin.ID, channel.ID)
	require.NoError(t, err)
}
