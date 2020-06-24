package vault

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNonce(t *testing.T) {
	var err error

	vlt := New(NewMem())
	err = vlt.checkNonce("123")
	require.NoError(t, err)
	err = vlt.commitNonce("123")
	require.NoError(t, err)
	err = vlt.checkNonce("123")
	require.EqualError(t, err, "nonce collision 123")
}
