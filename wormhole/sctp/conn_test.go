package sctp_test

import (
	"testing"

	"github.com/keys-pub/keys-ext/wormhole/sctp"
	"github.com/stretchr/testify/require"
)

func TestIsPrivateIP(t *testing.T) {
	require.True(t, sctp.IsPrivateIP("192.168.1.2"))
	require.False(t, sctp.IsPrivateIP("192.169.1.2"))
	require.True(t, sctp.IsPrivateIP("172.16.2.2"))
	require.True(t, sctp.IsPrivateIP("172.17.2.2"))
	require.False(t, sctp.IsPrivateIP("172.15.2.2"))
	require.True(t, sctp.IsPrivateIP("10.1.2.3"))
	require.True(t, sctp.IsPrivateIP("10.100.2.3"))
	require.False(t, sctp.IsPrivateIP("8.8.8.8"))
}
