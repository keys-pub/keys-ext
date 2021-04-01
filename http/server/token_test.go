package server_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokens(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))
	// firestore.SetContextLogger(NewContextLogger(DebugLevel))

	env := newEnv(t)
	serverEnv := newTestServerEnv(t, env)
	server := serverEnv.Server

	token, err := server.GenerateToken()
	require.NoError(t, err)
	require.Equal(t, 84, len(token))

	err = server.ValidateToken(token)
	require.NoError(t, err)
}
