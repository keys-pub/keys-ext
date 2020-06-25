package vault

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNonce(t *testing.T) {
	var err error

	vlt := New(NewMem())
	n := bytes.Repeat([]byte{0x01}, 24)
	err = vlt.checkNonce(n)
	require.NoError(t, err)
	err = vlt.commitNonce(n)
	require.NoError(t, err)
	err = vlt.checkNonce(n)
	require.EqualError(t, err, "nonce collision 00fdQWfEmi1CsDnkmh2kgfFBdcOWBGwvR")
}
