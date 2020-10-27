package matter_test

import (
	"context"
	"testing"

	"github.com/keys-pub/keys-ext/matter"
	"github.com/stretchr/testify/require"
)

func TestAdminUser(t *testing.T) {
	var err error
	ctx := context.TODO()

	matter.SetLogger(matter.NewLogger(matter.DebugLevel))

	client, err := matter.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	admin, err := client.LoginWithPassword(ctx, "gabriel", "testpassword")
	require.NoError(t, err)

	team, err := client.TeamByName(ctx, "test")
	require.NoError(t, err)
	require.NotNil(t, team)

	err = client.AddUserToTeam(ctx, admin.ID, team.ID)
	require.NoError(t, err)

	channel, err := client.ChannelByName(ctx, team.ID, "town-square")
	require.NoError(t, err)

	err = client.AddUserToChannel(ctx, admin.ID, channel.ID)
	require.NoError(t, err)
}
