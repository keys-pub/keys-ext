package matter_test

import (
	"context"
	"testing"

	"github.com/keys-pub/keys-ext/matter"
	"github.com/stretchr/testify/require"
)

func TestTeam(t *testing.T) {
	var err error

	client, err := matter.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	_, err = client.LoginWithPassword(context.TODO(), "gabriel", "testpassword")
	require.NoError(t, err)

	team, err := client.TeamByName(context.TODO(), "unknown")
	require.NoError(t, err)
	require.Nil(t, team)
}

func TestTeamCreate(t *testing.T) {
	var err error

	client, err := matter.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	_, err = client.LoginWithPassword(context.TODO(), "gabriel", "testpassword")
	require.NoError(t, err)

	team, err := client.CreateTeam(context.TODO(), "test", "Test", matter.TeamOpen)
	require.NoError(t, err)
	require.NotNil(t, team)
}
