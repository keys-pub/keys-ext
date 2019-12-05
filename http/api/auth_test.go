package api

import (
	"bytes"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/stretchr/testify/require"
)

const aliceSeed = "win rebuild update term layer transfer gain field prepare unique spider cool present argue grab trend eagle casino peace hockey loop seed desert swear"

func TestAuth(t *testing.T) {
	alice, err := keys.NewKeyFromSeedPhrase(aliceSeed, false)
	require.NoError(t, err)

	tm := keys.TimeFromMillis(123456789000)
	nonce := keys.Bytes32(bytes.Repeat([]byte{0x01}, 32))
	urs := "https://keys.pub" + keys.Path("direct", "message") + "?version=123456789001"
	auth, err := newAuth("GET", urs, tm, nonce, alice)
	require.NoError(t, err)
	require.Equal(t, "HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec:PgEJemfSCqHu06cPY7QwuioCMtl9P1YpmNfmzXL0WTzilvPXGBNkBp4q6gDrPmEO2csM0S3ozpKIMHFBPeBUTx", auth.Header())
	require.Equal(t, "https://keys.pub/direct/message?nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=123456789000&version=123456789001", auth.URL.String())

	req, err := newRequest("GET", urs, nil, tm, nonce, alice)
	require.NoError(t, err)
	require.Equal(t, "https://keys.pub/direct/message?nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=123456789000&version=123456789001", req.URL.String())
	require.Equal(t, "HX7DWqV9FtkXWJpXw656Uabtt98yjPH8iybGkfz2hvec:PgEJemfSCqHu06cPY7QwuioCMtl9P1YpmNfmzXL0WTzilvPXGBNkBp4q6gDrPmEO2csM0S3ozpKIMHFBPeBUTx", req.Header.Get("Authorization"))
}
