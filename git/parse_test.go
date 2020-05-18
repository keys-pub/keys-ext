package git_test

import (
	"testing"

	"github.com/keys-pub/keysd/git"
	"github.com/stretchr/testify/require"
)

func TestParseHost(t *testing.T) {
	host, err := git.ParseHost("git@gitlab.com:gabrielha/test.git")
	require.NoError(t, err)
	require.Equal(t, "gitlab.com", host)

	host, err = git.ParseHost("gitlab.com:gabrielha/test.git")
	require.NoError(t, err)
	require.Equal(t, "gitlab.com", host)

	host, err = git.ParseHost(":gabrielha/test.git")
	require.NoError(t, err)
	require.Equal(t, "", host)

	host, err = git.ParseHost("gabrielha/test.git")
	require.EqualError(t, err, "unrecognized git url format")

	host, err = git.ParseHost("git@gitlab.com")
	require.EqualError(t, err, "unrecognized git url format")

	host, err = git.ParseHost("gitlab.com")
	require.EqualError(t, err, "unrecognized git url format")
}
