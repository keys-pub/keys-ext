package client_test

import (
	"context"
	"testing"

	"github.com/keys-pub/keys-ext/matter/client"
	"github.com/stretchr/testify/require"
)

func TestTeam(t *testing.T) {
	var err error

	cl, err := client.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	_, err = cl.LoginWithPassword(context.TODO(), "gabriel", "testpassword")
	require.NoError(t, err)

	team, err := cl.TeamByName(context.TODO(), "unknown")
	require.NoError(t, err)
	require.Nil(t, team)
}

func TestTeamCreate(t *testing.T) {
	var err error

	cl, err := client.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	_, err = cl.LoginWithPassword(context.TODO(), "gabriel", "testpassword")
	require.NoError(t, err)

	team, err := cl.CreateTeam(context.TODO(), "test", "Test", client.TeamOpen)
	require.NoError(t, err)
	require.NotNil(t, team)
}
