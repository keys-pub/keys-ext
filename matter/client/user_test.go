package client_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/matter/client"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	var err error
	ctx := context.TODO()

	cl, err := client.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	username := keys.RandUsername(8)

	created, err := cl.CreateUser(ctx, &client.User{
		Username: username,
		Password: "password",
		Email:    fmt.Sprintf("%s@test.com", username),
	})
	require.NoError(t, err)
	require.NotNil(t, created)

	// Admin
	_, err = cl.LoginWithPassword(ctx, "gabriel", "testpassword")
	require.NoError(t, err)

	team, err := cl.TeamByName(ctx, "test")
	require.NoError(t, err)
	require.NotNil(t, team)

	err = cl.AddUserToTeam(ctx, created.ID, team.ID)
	require.NoError(t, err)

	channel, err := cl.ChannelByName(ctx, team.ID, "town-square")
	require.NoError(t, err)

	err = cl.AddUserToChannel(ctx, created.ID, channel.ID)
	require.NoError(t, err)

	// User
	_, err = cl.LoginWithPassword(ctx, username, "password")
	require.NoError(t, err)

	teams, err := cl.TeamsForUser(ctx, "")
	require.NoError(t, err)
	require.NotEmpty(t, len(teams))
	spew.Dump(teams)
}

func TestCreateUserWithKey(t *testing.T) {
	var err error
	ctx := context.TODO()

	cl, err := client.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	_, err = cl.LoginWithPassword(ctx, "gabriel", "testpassword")
	require.NoError(t, err)

	team, err := cl.TeamByName(ctx, "test")
	require.NoError(t, err)
	require.NotNil(t, team)

	key := keys.GenerateEdX25519Key()
	created, err := cl.CreateUserWithKey(ctx, key)
	require.NoError(t, err)

	err = cl.AddUserToTeam(ctx, created.ID, team.ID)
	require.NoError(t, err)

	channel, err := cl.ChannelByName(ctx, team.ID, "town-square")
	require.NoError(t, err)

	err = cl.AddUserToChannel(ctx, created.ID, channel.ID)
	require.NoError(t, err)

	cl.Logout()

	user, err := cl.LoginWithKey(ctx, key)
	require.NoError(t, err)
	require.Equal(t, created.ID, user.ID)

	channels, err := cl.ChannelsForUser(ctx, "", team.ID)
	require.NoError(t, err)
	require.True(t, len(channels) > 0)

	posts, err := cl.PostsForChannel(ctx, channel.ID)
	require.NoError(t, err)
	require.True(t, len(posts.Order) > 0)
	// spew.Dump(posts.Posts[posts.Order[0]])

	post, err := cl.CreatePost(ctx, channel.ID, "test message")
	require.NoError(t, err)
	require.NotNil(t, post)
	// t.Logf("post: %+v", post)
}

func TestTeamsForUser(t *testing.T) {
	var err error
	ctx := context.TODO()

	cl, err := client.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	user, err := cl.LoginWithPassword(ctx, "testuser2", "testuser2password")
	require.NoError(t, err)
	require.NotNil(t, user)

	teams, err := cl.TeamsForUser(ctx, "")
	require.NoError(t, err)
	require.NotEmpty(t, len(teams))
	spew.Dump(teams)
}

func TestTeams(t *testing.T) {
	var err error
	ctx := context.TODO()

	cl, err := client.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	// Admin
	admin, err := cl.LoginWithPassword(ctx, "gabriel", "testpassword")
	require.NoError(t, err)
	require.NotNil(t, admin)
	spew.Dump(admin)

	teams, err := cl.Teams(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, len(teams))
	spew.Dump(teams)
}
