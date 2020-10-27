package matter_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/matter"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	var err error
	ctx := context.TODO()

	matter.SetLogger(matter.NewLogger(matter.DebugLevel))

	client, err := matter.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	username := keys.RandUsername(8)

	created, err := client.CreateUser(ctx, &matter.User{
		Username: username,
		Password: "password",
		Email:    fmt.Sprintf("%s@test.com", username),
	})
	require.NoError(t, err)
	require.NotNil(t, created)

	// Admin
	_, err = client.LoginWithPassword(ctx, "gabriel", "testpassword")
	require.NoError(t, err)

	team, err := client.TeamByName(ctx, "test")
	require.NoError(t, err)
	require.NotNil(t, team)

	err = client.AddUserToTeam(ctx, created.ID, team.ID)
	require.NoError(t, err)

	channel, err := client.ChannelByName(ctx, team.ID, "town-square")
	require.NoError(t, err)

	err = client.AddUserToChannel(ctx, created.ID, channel.ID)
	require.NoError(t, err)

	// User
	_, err = client.LoginWithPassword(ctx, username, "password")
	require.NoError(t, err)

	teams, err := client.TeamsForUser(ctx, "")
	require.NoError(t, err)
	require.NotEmpty(t, len(teams))
	spew.Dump(teams)
}

func TestCreateUserWithKey(t *testing.T) {
	var err error
	ctx := context.TODO()

	matter.SetLogger(matter.NewLogger(matter.DebugLevel))

	client, err := matter.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	_, err = client.LoginWithPassword(ctx, "gabriel", "testpassword")
	require.NoError(t, err)

	team, err := client.TeamByName(ctx, "test")
	require.NoError(t, err)
	require.NotNil(t, team)

	key := keys.GenerateEdX25519Key()
	created, err := client.CreateUserWithKey(ctx, key)
	require.NoError(t, err)

	err = client.AddUserToTeam(ctx, created.ID, team.ID)
	require.NoError(t, err)

	channel, err := client.ChannelByName(ctx, team.ID, "town-square")
	require.NoError(t, err)

	err = client.AddUserToChannel(ctx, created.ID, channel.ID)
	require.NoError(t, err)

	client.Logout()

	user, err := client.LoginWithKey(ctx, key)
	require.NoError(t, err)
	require.Equal(t, created.ID, user.ID)

	channels, err := client.ChannelsForUser(ctx, "", team.ID)
	require.NoError(t, err)
	require.True(t, len(channels) > 0)

	posts, err := client.PostsForChannel(ctx, channel.ID)
	require.NoError(t, err)
	require.True(t, len(posts.Order) > 0)
	// spew.Dump(posts.Posts[posts.Order[0]])

	post, err := client.CreatePost(ctx, channel.ID, "test message")
	require.NoError(t, err)
	require.NotNil(t, post)
	// t.Logf("post: %+v", post)
}

func TestTeamsForUser(t *testing.T) {
	var err error
	ctx := context.TODO()

	matter.SetLogger(matter.NewLogger(matter.DebugLevel))

	client, err := matter.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	user, err := client.LoginWithPassword(ctx, "testuser2", "testuser2password")
	require.NoError(t, err)
	require.NotNil(t, user)

	teams, err := client.TeamsForUser(ctx, "")
	require.NoError(t, err)
	require.NotEmpty(t, len(teams))
	spew.Dump(teams)
}

func TestTeams(t *testing.T) {
	var err error
	ctx := context.TODO()

	matter.SetLogger(matter.NewLogger(matter.DebugLevel))

	client, err := matter.NewClient("http://localhost:8065/")
	require.NoError(t, err)

	// Admin
	admin, err := client.LoginWithPassword(ctx, "gabriel", "testpassword")
	require.NoError(t, err)
	require.NotNil(t, admin)
	spew.Dump(admin)

	teams, err := client.Teams(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, len(teams))
	spew.Dump(teams)
}
