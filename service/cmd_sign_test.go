package service

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignVerifyCommand(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))

	env := newTestEnv(t)
	service, closeFn := newTestService(t, env)
	defer closeFn()

	client, closeClFn := newTestRPCClient(t, service, env)
	defer closeClFn()

	testAuthSetup(t, service)
	testImportKey(t, service, alice)

	var clientErr error
	errorFn := func(err error) {
		clientErr = err
	}

	build := Build{Version: VersionDev}

	// Default (armored, detached) (file)
	inPath := writeTestFile(t)
	sigPath := inPath + ".sig"
	defer os.Remove(inPath)
	defer os.Remove(sigPath)

	cmd := append(os.Args[0:1], "-app", env.appName) // , "-log-level=debug")

	argsSign := append(cmd, "sign", "-s", alice.ID().String(), "-in", inPath)
	runClient(build, argsSign, client, errorFn)
	require.NoError(t, clientErr)
	require.FileExists(t, sigPath)
	sig, err := ioutil.ReadFile(sigPath)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(sig), "BEGIN SALTPACK DETACHED SIGNATURE."))

	argsVerify := append(cmd, "verify", "-in", inPath)
	runClient(build, argsVerify, client, errorFn)
	require.NoError(t, clientErr)

	// Binary, detached (file)
	inPath = writeTestFile(t)
	sigPath = inPath + ".sig"
	defer os.Remove(inPath)
	defer os.Remove(sigPath)

	argsSign = append(cmd, "sign", "-m", "binary", "-s", alice.ID().String(), "-in", inPath)
	runClient(build, argsSign, client, errorFn)
	require.NoError(t, clientErr)
	require.FileExists(t, sigPath)

	argsVerify = append(cmd, "verify", "-m", "binary", "-in", inPath)
	runClient(build, argsVerify, client, errorFn)
	require.NoError(t, clientErr)

	// Amrored, attached (file)
	inPath = writeTestFile(t)
	outPath := inPath + ".signed"
	defer os.Remove(inPath)
	defer os.Remove(sigPath)

	argsSign = append(cmd, "sign", "-m", "armor,attached", "-s", alice.ID().String(), "-in", inPath)
	runClient(build, argsSign, client, errorFn)
	require.NoError(t, clientErr)
	require.FileExists(t, outPath)
	signed, err := ioutil.ReadFile(outPath)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(signed), "BEGIN SALTPACK SIGNED MESSAGE."))
	os.Remove(inPath)

	argsVerify = append(cmd, "verify", "-m", "armor,attached", "-in", outPath)
	runClient(build, argsVerify, client, errorFn)
	require.NoError(t, clientErr)

	in, err := ioutil.ReadFile(inPath)
	require.NoError(t, err)
	require.Equal(t, string(in), "test message")

}
